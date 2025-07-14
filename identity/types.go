package identity

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 64
	MinUsernameLength = 3
	MaxUsernameLength = 64
)

var (
	UsernameMsg = fmt.Sprintf(
		"Username min length is %d and max length is %d",
		MinUsernameLength,
		MaxUsernameLength,
	)

	PasswordMsg = fmt.Sprintf(
		"Password min length is %d and max length is %d",
		MinPasswordLength,
		MaxPasswordLength,
	)
)

type PlainCredentials struct {
	Username string
	Password string
}

func (req *PlainCredentials) IsValid() (string, bool) {
	if n := len(req.Username); n < MinUsernameLength || n > MaxUsernameLength {
		return UsernameMsg, false
	}

	if n := len(req.Password); n < MinPasswordLength || n > MaxPasswordLength {
		return PasswordMsg, false
	}

	return "", true
}

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
	User          *U        `json:"user"`
	UserNotFound  bool      `json:"userNotFound"`
	NotAuthorized bool      `json:"notAuthorized"`
	Tokens        TokenPair `json:"tokens"`
}

type PlainSigninResult[U User] struct {
	User          *U              `json:"user"`
	UserNotFound  bool            `json:"userNotFound"`
	NotAuthorized bool            `json:"notAuthorized"`
	Tokens        StringTokenPair `json:"tokens"`
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
		User:          r.User,
		UserNotFound:  r.UserNotFound,
		NotAuthorized: r.NotAuthorized,
		Tokens:        r.Tokens.AsStringPair(),
	}
}
