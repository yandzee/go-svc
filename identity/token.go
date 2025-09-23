package identity

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Token struct {
	JWT       *jwt.Token
	JWTString string
}

func (t *Token) AsCookie(name string) http.Cookie {
	val := t.JWTString
	if len(val) == 0 {
		val = t.JWT.Raw
	}

	if t.JWT == nil {
		return http.Cookie{
			Name:  name,
			Value: val,
		}
	}

	remaining, _ := t.Remaining()
	maxAge := int(remaining.Seconds())
	if maxAge <= 0 {
		maxAge = -1
	}

	return http.Cookie{
		Name:     name,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
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

// Implements json.Marshaler
func (t *Token) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", t.RawString())

	return []byte(str), nil
}
