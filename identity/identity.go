package identity

import (
	"context"

	"github.com/google/uuid"
)

type Provider[U User] interface {
	SignIn(context.Context, *PlainCredentials) (*SigninResult[U], error)
	SignUp(context.Context, *PlainCredentials) (*SignupResult[U], error)
	Refresh(context.Context, *Token) (TokenPair, error)
}

type User interface {
	GetId() uuid.UUID
}
