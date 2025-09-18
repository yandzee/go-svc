package identity

import (
	"context"

	"github.com/google/uuid"
)

type Provider[U User] interface {
	SignIn(context.Context, SigninRequest) (*SigninResult[U], error)
	SignUp(context.Context, SignupRequest) (*SignupResult[U], error)
	Refresh(context.Context, *Token) (TokenPair, error)
	GetTokenUser(context.Context, *Token) (*U, error)
}

type UsersRegistry[U User] interface {
	CreateUser(context.Context, *UserStub) (CreateUserResult[U], error)
	GetUserById(context.Context, *uuid.UUID) (*U, error)
	GetUserByCredentials(context.Context, Credentials) (*U, error)
	CheckFieldsCorrectness(context.Context, Credentials) (CredentialsCheck, error)
	UserHasCredentials(context.Context, *U, Credentials) (bool, error)
}

type IdentityUtils interface {
	GenerateSalt() string
	Salt(string, string) string
}

type User interface {
	GetId() uuid.UUID
}
