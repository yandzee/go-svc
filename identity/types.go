package identity

import "context"

type Provider[User any] interface {
	// SignIn(context.Context, *SigninRequest) (*SigninResult, error)
	SignUp(context.Context, *SignupRequest) (*SignupResult[User], error)
}

type SigninRequest struct {
	Username string
	Password string
}

type SigninResult struct {
	IsInvalid bool
}
