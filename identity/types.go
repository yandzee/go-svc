package identity

import (
	"github.com/google/uuid"
)

type SignupResult[U User] struct {
	User          *U        `json:"user"`
	AlreadyExists bool      `json:"alreadyExists"`
	Tokens        TokenPair `json:"tokens"`
}

type PlainSignupResult[U any] struct {
	User          *U              `json:"user"`
	AlreadyExists bool            `json:"alreadyExists"`
	Tokens        StringTokenPair `json:"tokens"`
}

type SigninResult[U User] struct {
	User               *U        `json:"user"`
	UserNotFound       bool      `json:"userNotFound"`
	NotAuthorized      bool      `json:"notAuthorized"`
	InvalidCredentials bool      `json:"invalidCredentials"`
	Tokens             TokenPair `json:"tokens"`
}

type PlainSigninResult[U User] struct {
	User               *U              `json:"user"`
	UserNotFound       bool            `json:"userNotFound"`
	NotAuthorized      bool            `json:"notAuthorized"`
	InvalidCredentials bool            `json:"invalidCredentials"`
	Tokens             StringTokenPair `json:"tokens"`
}

type UserStub struct {
	Id           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	Salt         string    `json:"-"`
	PasswordHash string    `json:"-"`
}

func (r *SignupResult[U]) AsPlain() PlainSignupResult[U] {
	return PlainSignupResult[U]{
		User:          r.User,
		AlreadyExists: r.AlreadyExists,
		Tokens:        r.Tokens.AsStringPair(),
	}
}

func (r *SigninResult[U]) AsPlain() PlainSigninResult[U] {
	return PlainSigninResult[U]{
		User:               r.User,
		UserNotFound:       r.UserNotFound,
		NotAuthorized:      r.NotAuthorized,
		InvalidCredentials: r.InvalidCredentials,
		Tokens:             r.Tokens.AsStringPair(),
	}
}
