package identity

import "context"

type Provider interface {
	SignIn(context.Context, *SigninRequest) (*SigninResponse, error)
	SignUp(context.Context, *SignupRequest) (*SignupResponse, error)
}

type SigninRequest struct {
	Username string
	Password string
}

type SigninResult struct {
	IsInvalid bool
}
