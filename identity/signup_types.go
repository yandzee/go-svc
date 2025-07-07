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

type SignupRequest struct {
	Username string
	Password string
}

type UserStub struct {
	Id           uuid.UUID
	Username     string
	Password     string
	Salt         string
	PasswordHash string
}

func (req *SignupRequest) IsValid() (string, bool) {
	if n := len(req.Username); n < MinUsernameLength || n > MaxUsernameLength {
		return UsernameMsg, false
	}

	if n := len(req.Password); n < MinPasswordLength || n > MaxPasswordLength {
		return PasswordMsg, false
	}

	return "", true
}

type SignupResult[U any] struct {
	User   U
	Tokens TokenPair
}
