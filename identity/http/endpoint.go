package http

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandzee/go-svc/identity"
	"github.com/yandzee/go-svc/jwtutils"
	"github.com/yandzee/go-svc/log"
	"github.com/yandzee/go-svc/server/router"
)

const (
	AccessTokenHeader  = "X-Access-Token"
	RefreshTokenHeader = "X-Refresh-Token"
)

type IdentityEndpoint[U identity.User] struct {
	Provider           identity.Provider[U]
	Log                *slog.Logger
	AccessTokenHeader  string
	RefreshTokenHeader string
	TokenPrivateKey    *ecdsa.PrivateKey
}

func Wrap[U identity.User](id identity.Provider[U]) *IdentityEndpoint[U] {
	return &IdentityEndpoint[U]{
		Provider: id,
	}
}

func (ep *IdentityEndpoint[U]) Check() router.Handler {
	log := ep.log()

	return func(w http.ResponseWriter, r *http.Request, ctx router.Context) {
		pair, err := ep.tokensFromRequest(r)
		if err != nil {
			log.Error("tokensFromRequest failure", "err", err.Error())
			http.Error(
				w,
				"Auth check has failed: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		switch {
		case pair.AccessToken == nil:
			http.Error(w, "Unauthorized: no access token", http.StatusUnauthorized)
		case pair.AccessToken.Validation.IsExpired:
			http.Error(w, "Unauthorized: token is expired", http.StatusUnauthorized)
		case pair.AccessToken.Validation.IsMalformed:
			http.Error(w, "Unauthorized: token is malformed", http.StatusUnauthorized)
		case pair.AccessToken.Validation.IsParseError:
			err := pair.AccessToken.Validation.Error
			http.Error(w, "CheckAuth: token parse error: "+err.Error(), http.StatusInternalServerError)
		case pair.AccessToken.Validation.Error != nil:
			err := pair.AccessToken.Validation.Error
			http.Error(w, "CheckAuth: unexpected error: "+err.Error(), http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (ep *IdentityEndpoint[U]) Signup() router.Handler {
	log := ep.log()

	return func(w http.ResponseWriter, r *http.Request, ctx router.Context) {
		signupRequest := identity.PlainCredentials{}
		jsoner := ctx.Jsoner()

		res := jsoner.DecodeRequest(w, r, &signupRequest)
		if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
			log.Error("Signup body parse failure", "err", msg)
			http.Error(w, msg, st)
			return
		}

		log.Debug("Signup", "request", signupRequest)
		if msg, ok := signupRequest.IsValid(); !ok {
			log.Debug("Signup invalid credentials", "msg", msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		signupResult, err := ep.Provider.SignUp(r.Context(), &signupRequest)
		if err != nil {
			log.Error("Signup failed", "err", err.Error())
			http.Error(w, "Signup failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_ = jsoner.EncodeResponse(w, signupResult.AsPlain())
	}
}

func (ep *IdentityEndpoint[U]) Signin() router.Handler {
	log := ep.log()

	return func(w http.ResponseWriter, r *http.Request, ctx router.Context) {
		creds := identity.PlainCredentials{}
		jsoner := ctx.Jsoner()

		res := jsoner.DecodeRequest(w, r, &creds)
		if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
			log.Error("Signin body parse failure", "err", msg)
			http.Error(w, msg, st)
			return
		}

		log.Debug("Signin", "credentials", creds)
		if msg, ok := creds.IsValid(); !ok {
			log.Debug("Signin invalid credentials", "msg", msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		signinResult, err := ep.Provider.SignIn(r.Context(), &creds)
		if err != nil {
			log.Error("Signin failed", "err", err.Error())
			http.Error(w, "Signin failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_ = jsoner.EncodeResponse(w, signinResult.AsPlain())

		if signinResult.NotAuthorized {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func (ep *IdentityEndpoint[U]) Refresh() router.Handler {
	log := ep.log()

	return func(w http.ResponseWriter, r *http.Request, ctx router.Context) {
		pair, err := ep.tokensFromRequest(r)
		if err != nil {
			log.Error("Refresh failure", "err", err.Error())
			http.Error(
				w,
				"Refresh has failed: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		switch {
		case pair.RefreshToken == nil:
			http.Error(w, "Bad request: refresh token must be attached", http.StatusBadRequest)
		case pair.RefreshToken.Validation.IsExpired:
			http.Error(w, "Unauthorized: token is expired", http.StatusUnauthorized)
		case pair.RefreshToken.Validation.IsMalformed:
			http.Error(w, "Unauthorized: token is malformed", http.StatusUnauthorized)
		case pair.RefreshToken.Validation.IsParseError:
			err := pair.RefreshToken.Validation.Error
			http.Error(w, "RefreshAuth: token parse error"+err.Error(), http.StatusInternalServerError)
		case pair.RefreshToken.Validation.Error != nil:
			err := pair.RefreshToken.Validation.Error
			http.Error(w, "CheckAuth: unexpected error: "+err.Error(), http.StatusInternalServerError)
		}

		if !pair.RefreshToken.Validation.IsOk() {
			return
		}

		tokenPair, err := ep.Provider.Refresh(r.Context(), pair.RefreshToken.Token)
		if err != nil {
			http.Error(w, "Refresh: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := ep.respondWithTokenPair(w, &tokenPair); err != nil {
			http.Error(
				w,
				"Refresh: failed to respond with new tokens: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}
	}
}

func (ep *IdentityEndpoint[U]) respondWithTokenPair(w http.ResponseWriter, pair *identity.TokenPair) error {
	if token := pair.AccessToken; token != nil {
		w.Header().Set(ep.accessTokenHeaderName(), token.JWTString)
	}

	if token := pair.RefreshToken; token != nil {
		w.Header().Set(ep.refreshTokenHeaderName(), token.JWTString)
	}

	_, err := fmt.Fprintf(
		w,
		"Success: %d tokens, %s, have been placed to headers",
		pair.Num(),
		pair.Kinds(),
	)

	return err
}

func (ep *IdentityEndpoint[U]) tokensFromRequest(r *http.Request) (identity.ValidatedTokenPair, error) {
	accessTokenHeader := r.Header.Get(ep.accessTokenHeaderName())
	refreshTokenHeader := r.Header.Get(ep.refreshTokenHeaderName())

	pair := identity.ValidatedTokenPair{}
	var err error

	if len(accessTokenHeader) > 0 {
		pair.AccessToken, err = ep.parseToken(accessTokenHeader)
		if err != nil {
			return pair, errors.Join(
				fmt.Errorf("access token error"),
				err,
			)
		}
	}

	if len(refreshTokenHeader) > 0 {
		pair.RefreshToken, err = ep.parseToken(refreshTokenHeader)
		if err != nil {
			return pair, errors.Join(
				fmt.Errorf("refresh token error"),
				err,
			)
		}
	}

	return pair, nil
}

func (ep *IdentityEndpoint[U]) parseToken(tokenStr string) (*identity.ValidatedToken, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return &ep.TokenPrivateKey.PublicKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Name}),
	)

	tokenValidation, err := jwtutils.ValidateTokenParseError(err)
	if err != nil {
		return nil, err
	}

	return &identity.ValidatedToken{
		Token: &identity.Token{
			JWT:       token,
			JWTString: tokenStr,
		},
		Validation: tokenValidation,
	}, nil
}

func (ep *IdentityEndpoint[U]) accessTokenHeaderName() string {
	if len(ep.AccessTokenHeader) == 0 {
		return AccessTokenHeader
	}

	return ep.AccessTokenHeader
}

func (ep *IdentityEndpoint[U]) refreshTokenHeaderName() string {
	if len(ep.RefreshTokenHeader) == 0 {
		return RefreshTokenHeader
	}

	return ep.RefreshTokenHeader
}

func (ep *IdentityEndpoint[U]) log() *slog.Logger {
	return log.OrDiscard(ep.Log)
}
