package identity

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Token struct {
	JWT       *jwt.Token
	JWTString string
}

func (t *Token) RawString() string {
	if len(t.JWTString) > 0 {
		return t.JWTString
	}

	if t.JWT != nil {
		return t.JWT.Raw
	}

	return ""
}

func (t *Token) Remaining() (time.Duration, error) {
	numDate, err := t.JWT.Claims.GetExpirationTime()
	if err != nil {
		return 0, err
	}

	return time.Until(numDate.Time), nil
}

func (t *Token) GetUserId() (uuid.UUID, bool) {
	if t.JWT == nil {
		return uuid.Nil, false
	}

	subj, err := t.JWT.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, false
	}

	uid, err := uuid.Parse(subj)
	if err != nil {
		return uuid.Nil, false
	}

	return uid, true
}
