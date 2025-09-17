package identity

import "fmt"

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
	if msg, ok := req.IsValidUsername(); !ok {
		return msg, false
	}

	return req.IsValidPassword()
}

func (req *PlainCredentials) IsValidPassword() (string, bool) {
	if n := len(req.Password); n < MinPasswordLength || n > MaxPasswordLength {
		return PasswordMsg, false
	}

	return "", true

}

func (req *PlainCredentials) IsValidUsername() (string, bool) {
	if n := len(req.Username); n < MinUsernameLength || n > MaxUsernameLength {
		return UsernameMsg, false
	}

	return "", true
}
