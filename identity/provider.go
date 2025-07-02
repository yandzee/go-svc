package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/yandzee/go-svc/crypto"
	"github.com/yandzee/gotx"
)

type RegistryProvider[U any] struct {
	Log *slog.Logger
	// Txer     *gotx.AnyTransactor
	Registry UsersRegistry[U]
}

type UsersRegistry[U any] interface {
	CreateUser(context.Context, &UserStub) (U, error)
}

func (p *RegistryProvider) Signin(ctx context.Context, r *SigninRequest) (*SigninResult, error) {
	return nil, nil
}

func (p *RegistryProvider) Signup(
	ctx context.Context,
	req *SignupRequest,
) (*SignupResult, error) {
	if req == nil {
		return nil, errors.New("Cannot signup using nil request")
	}

	if p.Registry == nil {
		return nil, errors.New("Cannot signup using nil UsersRegistry")
	}

	if _, ok := req.IsValid(); !ok {
		return nil, errors.New("Cannot signup using invalid credentials")
	}

	// tx, err := p.Txer.Context(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	userId, err := uuid.NewV7()
	if err != nil {
		// return nil, errors.Join(err, tx.Rollback(ctx))
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

	err = p.Registry.CreateUser(ctx, &stub)
	if err != nil {
		return nil, err
	}

	tokenPair, err := svc.createSignedTokenPair(&existing.Id)
	if err != nil {
		svc.Log.Error("issueJWTString failure", "err", err.Error())
		return nil, err
	}

	return &auth.SignupResult{
		User:            existing,
		Tokens:          tokenPair.AsStringPair(),
		IsUsernameTaken: false,
	}, tx.Commit(ctx)
}

func (p *RegistryProvider) salt(smth string) (string, string) {
	salt := crypto.RandomSha256(32)

	h := sha256.New()
	fmt.Fprintf(h, "%s.%s", salt, smth)

	return salt, hex.EncodeToString(h.Sum(nil))
}
