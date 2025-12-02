package jwtutils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type TokenValidation struct {
	Error error

	IsExpired    bool
	IsMalformed  bool
	IsParseError bool
}

func ValidateTokenParseError(err error) (TokenValidation, error) {
	tv := TokenValidation{
		Error: err,
	}

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		tv.IsExpired = true
	case errors.Is(err, jwt.ErrInvalidType):
		tv.IsParseError = true
	case errors.Is(err, jwt.ErrInvalidKey):
		fallthrough
	case errors.Is(err, jwt.ErrInvalidKeyType):
		fallthrough
	case errors.Is(err, jwt.ErrHashUnavailable):
		return tv, err
	case errors.Is(err, jwt.ErrTokenMalformed):
		fallthrough
	case errors.Is(err, jwt.ErrTokenUnverifiable):
		fallthrough
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		fallthrough
	case errors.Is(err, jwt.ErrTokenRequiredClaimMissing):
		fallthrough
	case errors.Is(err, jwt.ErrTokenInvalidAudience):
		fallthrough
	case errors.Is(err, jwt.ErrTokenUsedBeforeIssued):
		fallthrough
	case errors.Is(err, jwt.ErrTokenInvalidIssuer):
		fallthrough
	case errors.Is(err, jwt.ErrTokenInvalidSubject):
		fallthrough
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		fallthrough
	case errors.Is(err, jwt.ErrTokenInvalidId):
		fallthrough
	case errors.Is(err, jwt.ErrTokenInvalidClaims):
		tv.IsMalformed = true
	}

	return tv, nil
}

func (tv *TokenValidation) IsOk() bool {
	return tv.Error == nil && !tv.IsExpired && !tv.IsMalformed && !tv.IsParseError
}
