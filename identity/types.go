package identity

import (
	"github.com/google/uuid"
)

type SignupRequest struct {
	Credentials Credentials `json:"credentials"`
}

type SigninRequest SignupRequest

type SignupResult[U User] struct {
	User               *U               `json:"user"`
	AlreadyExists      bool             `json:"alreadyExists"`
	InvalidCredentials bool             `json:"invalidCredentials"`
	CredentialsCheck   CredentialsCheck `json:"credentialsCheck"`
	Tokens             TokenPair        `json:"tokens"`
}

type SigninResult[U User] struct {
	User                *U               `json:"user"`
	NotAuthorized       bool             `json:"notAuthorized"`
	UserNotFound        bool             `json:"userNotFound"`
	CredentialsMismatch bool             `json:"credentialsMismatch"`
	InvalidCredentials  bool             `json:"invalidCredentials"`
	CredentialsCheck    CredentialsCheck `json:"credentialsCheck"`
	Tokens              TokenPair        `json:"tokens"`
}

type UserStub struct {
	Id          uuid.UUID   `json:"id"`
	Credentials Credentials `json:"credentials"`
}

func (r *SignupResult[U]) IsSuccess() bool {
	return r.User != nil && !r.AlreadyExists && !r.InvalidCredentials
}
