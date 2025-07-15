package identity

import (
	"github.com/google/uuid"
	"github.com/yandzee/go-svc/jwtutils"
)

type ValidatedToken struct {
	Token      *Token
	Validation jwtutils.TokenValidation
}

type ValidatedTokenPair struct {
	AccessToken  *ValidatedToken
	RefreshToken *ValidatedToken
}

func (vtp *ValidatedTokenPair) HasValidAccess() bool {
	return vtp.AccessToken != nil && vtp.AccessToken.IsValid()
}

func (vt *ValidatedToken) IsValid() bool {
	return vt.Token != nil && vt.Validation.IsOk()
}

func (vtp *ValidatedTokenPair) UserId() (uuid.UUID, bool) {
	if vtp.AccessToken != nil {
		uid, ok := vtp.AccessToken.Token.GetUserId()
		if ok {
			return uid, true
		}
	}

	if vtp.RefreshToken != nil {
		uid, ok := vtp.RefreshToken.Token.GetUserId()
		if ok {
			return uid, true
		}
	}

	return uuid.Nil, false
}
