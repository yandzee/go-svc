package identity

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yandzee/go-svc/log"
)

type RegistryProvider[U User] struct {
	Log      *slog.Logger
	Core     IdentityCore
	Registry UsersRegistry[U]

	BaseClaims           jwt.RegisteredClaims
	TokenPrivateKey      *ecdsa.PrivateKey
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type UsersRegistry[U User] interface {
	CreateUser(context.Context, *UserStub) (CreateUserResult[U], error)
	GetUserByUsername(context.Context, string) (*U, error)
	UserHasCredentials(
		context.Context,
		IdentityCore,
		*U,
		*PlainCredentials,
	) (CredsCheckResult, error)
}

type CreateUserResult[U User] struct {
	AlreadyExists bool
	User          *U
}

type CredsCheckResult struct {
	IsWrongPassword bool
}

func (p *RegistryProvider[U]) SignIn(
	ctx context.Context,
	creds *PlainCredentials,
) (*SigninResult[U], error) {
	if creds == nil {
		return nil, errors.New("cannot signin using nil credentials")
	}

	if _, ok := creds.IsValid(); !ok {
		return nil, errors.New("cannot signin using invalid credentials")
	}

	usr, err := p.Registry.GetUserByUsername(ctx, creds.Username)
	if err != nil {
		return nil, err
	}

	if usr == nil {
		return &SigninResult[U]{
			UserNotFound: true,
		}, nil
	}

	authCheck, err := p.Registry.UserHasCredentials(ctx, p.ensureCore(), usr, creds)
	if err != nil {
		return nil, err
	}

	if authCheck.IsWrongPassword {
		return &SigninResult[U]{
			NotAuthorized: true,
		}, nil
	}

	uid := (*usr).GetId()

	tokenPair, err := p.createSignedTokenPair(&uid)
	if err != nil {
		p.log().Error("createSignedTokenPair failure", "err", err.Error())
		return nil, err
	}

	return &SigninResult[U]{
		User:   usr,
		Tokens: tokenPair,
	}, nil
}

func (p *RegistryProvider[U]) SignUp(
	ctx context.Context,
	creds *PlainCredentials,
) (*SignupResult[U], error) {
	if creds == nil {
		return nil, errors.New("cannot signup using nil request")
	}

	if p.Registry == nil {
		return nil, errors.New("cannot signup using nil UsersRegistry")
	}

	if _, ok := creds.IsValid(); !ok {
		return nil, errors.New("cannot signup using invalid credentials")
	}

	userId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	core := p.ensureCore()

	salt := core.GenerateSalt()
	pwdHash := core.Salt(salt, creds.Password)

	stub := UserStub{
		Id:           userId,
		Username:     creds.Username,
		Salt:         salt,
		Password:     creds.Password,
		PasswordHash: pwdHash,
	}

	createResult, err := p.Registry.CreateUser(ctx, &stub)
	if err != nil {
		return nil, err
	}

	if createResult.AlreadyExists {
		return &SignupResult[U]{
			AlreadyExists: true,
		}, nil
	}

	tokenPair, err := p.createSignedTokenPair(&userId)
	if err != nil {
		p.log().Error("issueJWTString failure", "err", err.Error())
		return nil, err
	}

	return &SignupResult[U]{
		User:          createResult.User,
		Tokens:        tokenPair,
		AlreadyExists: false,
	}, nil
}

func (p *RegistryProvider[U]) Refresh(
	ctx context.Context,
	refreshToken *Token,
) (TokenPair, error) {
	if refreshToken == nil {
		return TokenPair{}, errors.New("refresh token is nil")
	}

	userId, isOk := refreshToken.GetUserId()
	if !isOk {
		return TokenPair{}, errors.New("refresh token contains")
	}

	return p.createSignedTokenPair(&userId)
}

func (p *RegistryProvider[U]) createSignedTokenPair(userId *uuid.UUID) (TokenPair, error) {
	var pair TokenPair
	var err error

	pair.AccessToken, err = p.createSignedToken(userId, "at", p.AccessTokenDuration)
	if err != nil {
		return pair, err
	}

	pair.RefreshToken, err = p.createSignedToken(userId, "rt", p.RefreshTokenDuration)
	if err != nil {
		return pair, err
	}

	return pair, err
}

func (p *RegistryProvider[U]) createSignedToken(
	userId *uuid.UUID,
	idPrefix string,
	dur time.Duration,
) (*Token, error) {
	if p.TokenPrivateKey == nil {
		return nil, errors.New("cannot create signed token: private key is nil")
	}

	if userId == nil {
		return nil, errors.New("cannot create signed token: userId is nil")
	}

	now := time.Now()
	uid := userId.String()

	claims := p.mergeClaims(jwt.RegisteredClaims{
		Subject:   uid,
		ExpiresAt: jwt.NewNumericDate(now.Add(dur)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        fmt.Sprintf("%s-%s-%d", idPrefix, uid, now.UnixNano()),
	})

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	signedTokenStr, err := token.SignedString(p.TokenPrivateKey)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to create signed token string"),
			err,
		)
	}

	return &Token{
		JWT:       token,
		JWTString: signedTokenStr,
	}, nil
}

func (p *RegistryProvider[U]) mergeClaims(filler jwt.RegisteredClaims) jwt.RegisteredClaims {
	if len(filler.Issuer) == 0 {
		filler.Issuer = p.BaseClaims.Issuer
	}

	filler.Audience = append(filler.Audience, p.BaseClaims.Audience...)

	return filler
}

func (p *RegistryProvider[U]) log() *slog.Logger {
	return log.OrDiscard(p.Log)
}

func (p *RegistryProvider[U]) ensureCore() IdentityCore {
	if p.Core != nil {
		return p.Core
	}

	p.Core = &DefaultCore{}
	return p.Core
}
