package http

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandzee/go-svc/data/jsoner"
	"github.com/yandzee/go-svc/identity"
	"github.com/yandzee/go-svc/jwtutils"
	"github.com/yandzee/go-svc/log"
	"github.com/yandzee/go-svc/router"
)

const (
	AccessTokenHeader  = "X-Access-Token"
	RefreshTokenHeader = "X-Refresh-Token"

	KiloByte = 1024
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

	return func(rctx *router.RequestContext) {
		pair, err := ep.tokensFromRequest(rctx.Request)
		if err != nil {
			log.Error("tokensFromRequest failure", "err", err.Error())

			rctx.Response.String(
				http.StatusInternalServerError,
				"Auth check has failed: "+err.Error(),
			)
			return
		}

		switch {
		case pair.AccessToken == nil:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: no access token")
		case pair.AccessToken.Validation.IsExpired:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: token is expired")
		case pair.AccessToken.Validation.IsMalformed:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: token is malformed")
		case pair.AccessToken.Validation.IsParseError:
			err := pair.AccessToken.Validation.Error
			rctx.Response.String(
				http.StatusInternalServerError,
				"CheckAuth: token parse error: "+err.Error(),
			)
		case pair.AccessToken.Validation.Error != nil:
			err := pair.AccessToken.Validation.Error
			rctx.Response.String(
				http.StatusInternalServerError,
				"CheckAuth: unexpected error: "+err.Error(),
			)
		default:
			rctx.Response.String(http.StatusOK)
			dur, err := pair.AccessToken.Token.Remaining()
			if err == nil {
				rctx.Response.Stringf(
					http.StatusOK,
					"CheckAuth: token is valid for duration: %s", dur,
				)
			}
		}
	}
}

func (ep *IdentityEndpoint[U]) Signup() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		signupRequest := identity.PlainCredentials{}
		jsoner := jsoner.Jsoner{}

		res := jsoner.Decode(rctx.Request.LimitedBody(16*KiloByte), &signupRequest)
		if err := res.Err(); err != nil {
			log.Error("Signup body parse failure", "err", err.Error())

			rctx.Response.Stringf(
				http.StatusBadRequest,
				"Failed to parse Signup data: %s",
				err.Error(),
			)
			return
		}

		log.Debug("Signup", "request", signupRequest)
		if msg, ok := signupRequest.IsValid(); !ok {
			log.Debug("Signup invalid credentials", "msg", msg)
			rctx.Response.String(http.StatusBadRequest, msg)
			return
		}

		signupResult, err := ep.Provider.SignUp(rctx.Context(), &signupRequest)
		if err != nil {
			log.Error("Signup failed", "err", err.Error())
			rctx.Response.Stringf(http.StatusInternalServerError, "Signup failure: %s", err.Error())
			return
		}

		_ = jsoner.Encode(rctx.Response, signupResult.AsPlain())
	}
}

func (ep *IdentityEndpoint[U]) Signin() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		creds := identity.PlainCredentials{}
		jsoner := jsoner.Jsoner{}

		res := jsoner.Decode(rctx.Request.LimitedBody(16*KiloByte), &creds)
		if err := res.Err(); err != nil {
			log.Error("Signin body parse failure", "err", err.Error())
			rctx.Response.Stringf(
				http.StatusBadRequest,
				"Failed to parse signin request: %s",
				err.Error(),
			)

			return
		}

		log.Debug("Signin", "credentials", creds)
		if msg, ok := creds.IsValid(); !ok {
			log.Debug("Signin invalid credentials", "msg", msg)
			rctx.Response.Stringf(http.StatusBadRequest, "Invalid signin credentials: %s", msg)
			return
		}

		signinResult, err := ep.Provider.SignIn(rctx.Context(), &creds)
		if err != nil {
			log.Error("Signin failed", "err", err.Error())
			rctx.Response.Stringf(http.StatusInternalServerError, "Signin failed: %s", err.Error())
			return
		}

		_ = jsoner.Encode(rctx.Response, signinResult.AsPlain())

		if signinResult.NotAuthorized {
			rctx.Response.String(http.StatusUnauthorized)
		}
	}
}

func (ep *IdentityEndpoint[U]) Refresh() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		pair, err := ep.tokensFromRequest(rctx.Request)
		if err != nil {
			log.Error("Refresh failure", "err", err.Error())
			rctx.Response.Stringf(
				http.StatusInternalServerError,
				"Refresh has failed: %s",
				err.Error(),
			)
			return
		}

		switch {
		case pair.RefreshToken == nil:
			rctx.Response.String(http.StatusBadRequest, "Refresh token must be attached")
		case pair.RefreshToken.Validation.IsExpired:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: token is expired")
		case pair.RefreshToken.Validation.IsMalformed:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: token is malformed")
		case pair.RefreshToken.Validation.IsParseError:
			err := pair.RefreshToken.Validation.Error
			rctx.Response.Stringf(http.StatusInternalServerError, "RefreshAuth: token parse error: %s", err.Error())
		case pair.RefreshToken.Validation.Error != nil:
			err := pair.RefreshToken.Validation.Error
			rctx.Response.Stringf(http.StatusInternalServerError, "CheckAuth: unexpected error: %s", err.Error())
		}

		if pair.RefreshToken == nil || !pair.RefreshToken.Validation.IsOk() {
			return
		}

		tokenPair, err := ep.Provider.Refresh(rctx.Context(), pair.RefreshToken.Token)
		if err != nil {
			rctx.Response.Stringf(http.StatusInternalServerError, "Refresh: %s", err.Error())
			return
		}

		if err := ep.respondWithTokenPair(rctx, &tokenPair); err != nil {
			rctx.Response.Stringf(
				http.StatusInternalServerError,
				"Refresh: failed to respond with new tokens: %s",
				err.Error(),
			)
			return
		}
	}
}

func (ep *IdentityEndpoint[U]) respondWithTokenPair(
	rctx *router.RequestContext,
	pair *identity.TokenPair,
) error {
	headers := rctx.Response.Headers()

	if token := pair.AccessToken; token != nil {
		headers.Set(ep.accessTokenHeaderName(), token.JWTString)
	}

	if token := pair.RefreshToken; token != nil {
		headers.Set(ep.refreshTokenHeaderName(), token.JWTString)
	}

	_, err := fmt.Fprintf(
		rctx.Response,
		"Success: %d tokens, %s, have been placed to headers",
		pair.Num(),
		pair.Kinds(),
	)

	return err
}

func (ep *IdentityEndpoint[U]) tokensFromRequest(r router.Request) (identity.ValidatedTokenPair, error) {
	headers := r.Headers()

	accessTokenHeader := headers.Get(ep.accessTokenHeaderName())
	refreshTokenHeader := headers.Get(ep.refreshTokenHeaderName())

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
