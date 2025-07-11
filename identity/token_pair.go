package identity

import "github.com/yandzee/gou/boolean"

type TokenPair struct {
	AccessToken  *Token `json:"accessToken"`
	RefreshToken *Token `json:"refreshToken"`
}

type StringTokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (tp *TokenPair) AsStringPair() StringTokenPair {
	pair := StringTokenPair{}

	if tp.AccessToken != nil {
		pair.AccessToken = tp.AccessToken.RawString()
	}

	if tp.RefreshToken != nil {
		pair.RefreshToken = tp.RefreshToken.RawString()
	}

	return pair
}

func (tp *TokenPair) Num() int {
	return boolean.ToInt(tp.AccessToken != nil) + boolean.ToInt(tp.RefreshToken != nil)
}

func (p *TokenPair) Kinds() string {
	switch {
	case p.AccessToken != nil && p.RefreshToken != nil:
		return "Access Token and Refresh Token"
	case p.AccessToken != nil:
		return "Access Token"
	case p.RefreshToken != nil:
		return "Refresh Token"
	}

	return "No tokens"
}
