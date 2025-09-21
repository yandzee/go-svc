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
	Registry UsersRegistry[U]

	BaseClaims           jwt.RegisteredClaims
	TokenPrivateKey      *ecdsa.PrivateKey
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type CreateUserResult[U User] struct {
	AlreadyExists bool
	User          *U
}

func (p *RegistryProvider[U]) SignIn(
	ctx context.Context,
	req SigninRequest,
) (*SigninResult[U], error) {
	if req.Credentials == nil {
		return nil, ErrNoCredentials
	}

	usr, err := p.Registry.GetUserByCredentials(ctx, req.Credentials)
	if err != nil {
		return nil, err
	}

	if usr == nil {
		return &SigninResult[U]{
			NotAuthorized: true,
			UserNotFound:  true,
		}, nil
	}

	has, err := p.Registry.UserHasCredentials(ctx, usr, req.Credentials)
	if err != nil {
		return nil, err
	}

	if !has {
		return &SigninResult[U]{
			NotAuthorized:      true,
			InvalidCredentials: true,
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
	req SignupRequest,
) (*SignupResult[U], error) {
	if req.Credentials == nil {
		return nil, ErrNoCredentials
	}

	if p.Registry == nil {
		return nil, errors.New("cannot signup using nil UsersRegistry")
	}

	if ch, err := p.Registry.CheckFieldsCorrectness(ctx, req.Credentials); err != nil {
		return nil, err
	} else if _, has := ch.HasIncorrect(); has {
		return &SignupResult[U]{
			InvalidCredentials: true,
			CredentialsCheck:   ch,
		}, nil
	}

	userId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	stub := UserStub{
		Id:          userId,
		Credentials: req.Credentials,
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
		return TokenPair{}, errors.New("refresh token contains invalid user id")
	}

	return p.createSignedTokenPair(&userId)
}

func (p *RegistryProvider[U]) GetTokenUser(
	ctx context.Context,
	token *Token,
) (*U, error) {
	if token == nil {
		return nil, errors.New("token is nil")
	}

	userId, isOk := token.GetUserId()
	if !isOk {
		return nil, errors.New("token contains invalid user id")
	}

	return p.Registry.GetUserById(ctx, &userId)
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
