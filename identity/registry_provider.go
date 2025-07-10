package identity

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yandzee/go-svc/crypto"
	"github.com/yandzee/go-svc/log"
)

type RegistryProvider[U any] struct {
	Log      *slog.Logger
	Registry UsersRegistry[U]

	BaseClaims           jwt.RegisteredClaims
	TokenPrivateKey      *ecdsa.PrivateKey
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type UsersRegistry[U any] interface {
	CreateUser(context.Context, *UserStub) (CreateUserResult[U], error)
}

type CreateUserResult[U any] struct {
	AlreadyExists bool
	User          *U
}

func (p *RegistryProvider[U]) Signin(ctx context.Context, r *SigninRequest) (*SigninResult, error) {
	return nil, nil
}

func (p *RegistryProvider[U]) SignUp(
	ctx context.Context,
	req *SignupRequest,
) (*SignupResult[U], error) {
	if req == nil {
		return nil, errors.New("cannot signup using nil request")
	}

	if p.Registry == nil {
		return nil, errors.New("cannot signup using nil UsersRegistry")
	}

	if _, ok := req.IsValid(); !ok {
		return nil, errors.New("cannot signup using invalid credentials")
	}

	userId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	salt, pwdHash := p.salt(req.Password)
	stub := UserStub{
		Id:           userId,
		Username:     req.Username,
		Salt:         salt,
		Password:     req.Password,
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

func (p *RegistryProvider[U]) salt(smth string) (string, string) {
	salt := crypto.RandomSha256(32)

	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s.%s", salt, smth)

	return salt, hex.EncodeToString(h.Sum(nil))
}

func (p *RegistryProvider[U]) log() *slog.Logger {
	return log.OrDiscard(p.Log)
}
