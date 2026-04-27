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
	"github.com/yandzee/go-svc/log"
	"github.com/yandzee/go-svc/router"
	jwtutils "github.com/yandzee/go-svc/utils/jwt"
)

const (
	DefaultAccessTokenKey  = "X-Access-Token"
	DefaultRefreshTokenKey = "X-Refresh-Token"

	KiloByte = 1024
)

type AuthCredentialsMedia int

const (
	// NOTE: Auth credentials are managed via unsecured plain http headers
	PlainHeadersMedia AuthCredentialsMedia = iota

	// NOTE: Auth credentials are managed via cookies
	CookiesMedia

	// NOTE: Same as CookieMedia but Secure flag is enabled
	SecureCookiesMedia
)

type IdentityEndpoint[U identity.User] struct {
	Provider        identity.Provider[U]
	Log             *slog.Logger
	TokenPrivateKey *ecdsa.PrivateKey
	Media           AuthCredentialsMedia

	// NOTE: Those are strings which will be used to name headers / cookies
	AccessTokenKey  string
	RefreshTokenKey string

	// NOTE: Set this path to limit requests with refresh token attached as cookie
	RefreshTokenCookiePath string

	DisableHttpOnly bool
}

func Wrap[U identity.User](id identity.Provider[U]) *IdentityEndpoint[U] {
	return &IdentityEndpoint[U]{
		Provider: id,
	}
}

func (ep *IdentityEndpoint[U]) Check() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		atoken, err := ep.getAccessTokenFromRequest(rctx.Request)
		if err != nil {
			log.Error("tokensFromRequest failure", "err", err.Error())

			rctx.Response.String(
				http.StatusInternalServerError,
				"Auth check has failed: "+err.Error(),
			)
			return
		}

		switch {
		case atoken == nil:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: no access token")
		case atoken.Validation.IsExpired:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: token is expired")
		case atoken.Validation.IsMalformed:
			rctx.Response.String(http.StatusUnauthorized, "Unauthorized: token is malformed")
		case atoken.Validation.IsParseError:
			err := atoken.Validation.Error
			rctx.Response.String(
				http.StatusInternalServerError,
				"CheckAuth: token parse error: "+err.Error(),
			)
		case atoken.Validation.Error != nil:
			err := atoken.Validation.Error
			rctx.Response.String(
				http.StatusInternalServerError,
				"CheckAuth: unexpected error: "+err.Error(),
			)
		default:
			dur, _ := atoken.Token.Remaining()

			rctx.Response.Stringf(
				http.StatusOK,
				"CheckAuth: token is valid for duration: %s", dur,
			)
		}
	}
}

func (ep *IdentityEndpoint[U]) CurrentUser() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		result, err := ep.Guard(rctx, GuardOptions{
			IsOptional:          false,
			IsUserFetchDisabled: false,
		})

		if err != nil {
			if result.IsResponded {
				return
			}

			log.Error("CurrentUser", "err", err.Error())

			rctx.Response.String(
				http.StatusInternalServerError,
				"Failed to get current authorization: "+err.Error(),
			)

			return
		}

		if result.IsResponded {
			return
		}

		if _, err := rctx.Response.JSON(http.StatusOK, result.User); err != nil {
			log.Error("Failed to respond with user's json", "err", err.Error())

			rctx.Response.String(
				http.StatusInternalServerError,
				err.Error(),
			)
		}
	}
}

func (ep *IdentityEndpoint[U]) Signup() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		signupRequest := identity.SignupRequest{}
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

		signupResult, err := ep.Provider.SignUp(rctx.Context(), signupRequest)
		if err != nil {
			log.Error("Signup failed", "err", err.Error())
			rctx.Response.Stringf(http.StatusInternalServerError, "Signup failure: %s", err.Error())
			return
		}

		if signupResult.IsSuccess() {
			ep.setTokensToResponse(rctx, signupResult.Tokens.AccessToken, signupResult.Tokens.RefreshToken)
		}

		_, _ = rctx.Response.JSON(http.StatusOK, signupResult)
	}
}

func (ep *IdentityEndpoint[U]) Signin() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		signinRequest := identity.SigninRequest{}
		jsoner := jsoner.Jsoner{}

		res := jsoner.Decode(rctx.Request.LimitedBody(16*KiloByte), &signinRequest)
		if err := res.Err(); err != nil {
			log.Error("Signin body parse failure", "err", err.Error())
			rctx.Response.Stringf(
				http.StatusBadRequest,
				"Failed to parse signin request: %s",
				err.Error(),
			)

			return
		}

		log.Debug("Signin", "signinRequest", signinRequest)

		signinResult, err := ep.Provider.SignIn(rctx.Context(), signinRequest)
		switch {
		case errors.Is(err, identity.ErrNoCredentials):
			rctx.Response.String(http.StatusBadRequest, "Signin failed: no credentials provided")
			return
		case err != nil:
			log.Error("Signin failed", "err", err.Error())
			rctx.Response.Stringf(http.StatusInternalServerError, "Signin failed: %s", err.Error())
			return
		}

		log.Debug("Signin result", "result", signinResult)

		st := http.StatusOK
		if signinResult.NotAuthorized {
			st = http.StatusUnauthorized
		} else {
			ep.setTokensToResponse(rctx, signinResult.Tokens.AccessToken, signinResult.Tokens.RefreshToken)
		}

		_, _ = rctx.Response.JSON(st, signinResult)
	}
}

func (ep *IdentityEndpoint[U]) Refresh() router.Handler {
	log := ep.log()

	return func(rctx *router.RequestContext) {
		pair, err := ep.getTokensFromRequest(rctx.Request)
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

		ep.setTokensToResponse(rctx, tokenPair.AccessToken, tokenPair.RefreshToken)

		_, err = fmt.Fprintf(
			rctx.Response,
			"Success: %d tokens, %s, have been placed to headers",
			tokenPair.Num(),
			tokenPair.Kinds(),
		)

		if err != nil {
			rctx.Response.Stringf(
				http.StatusInternalServerError,
				"Refresh: failed to respond with new tokens: %s",
				err.Error(),
			)
			return
		}
	}
}

func (ep *IdentityEndpoint[U]) setTokensToResponse(
	rc *router.RequestContext,
	atoken *identity.Token,
	rtoken *identity.Token,
) {
	switch ep.Media {
	case PlainHeadersMedia:
		h := rc.Response.Headers()

		if atoken != nil {
			h.Set(ep.accessTokenKey(), atoken.JWTString)
		}

		if rtoken != nil {
			h.Set(ep.refreshTokenKey(), rtoken.JWTString)
		}
	default:
		if atoken != nil {
			cookie := atoken.AsCookie(ep.accessTokenKey())
			cookie.HttpOnly = !ep.DisableHttpOnly
			cookie.Secure = ep.Media == SecureCookiesMedia

			rc.Response.SetCookie(&cookie)
		}

		if rtoken != nil {
			cookie := rtoken.AsCookie(ep.refreshTokenKey())
			cookie.HttpOnly = !ep.DisableHttpOnly
			cookie.Secure = ep.Media == SecureCookiesMedia
			cookie.Path = ep.RefreshTokenCookiePath

			rc.Response.SetCookie(&cookie)
		}
	}
}

func (ep *IdentityEndpoint[U]) getTokensFromRequest(r router.Request) (identity.ValidatedTokenPair, error) {
	pair := identity.ValidatedTokenPair{}

	atoken, err := ep.getAccessTokenFromRequest(r)
	if err != nil {
		return pair, err
	}

	rtoken, err := ep.getRefreshTokenFromRequest(r)
	if err != nil {
		return pair, err
	}

	pair.AccessToken = atoken
	pair.RefreshToken = rtoken

	return pair, nil
}

func (ep *IdentityEndpoint[U]) getAccessTokenFromRequest(r router.Request) (*identity.ValidatedToken, error) {
	accessTokenString := ep.getAccessTokenStringFromRequest(r)

	if len(accessTokenString) == 0 {
		return nil, nil
	}

	atoken, err := ep.parseToken(accessTokenString)
	if err != nil {
		return atoken, errors.Join(
			fmt.Errorf("access token error"),
			err,
		)
	}

	return atoken, nil
}

func (ep *IdentityEndpoint[U]) getRefreshTokenFromRequest(r router.Request) (*identity.ValidatedToken, error) {
	refreshTokenString := ep.getRefreshTokenStringFromRequest(r)

	if len(refreshTokenString) == 0 {
		return nil, nil
	}

	rtoken, err := ep.parseToken(refreshTokenString)
	if err != nil {
		return rtoken, errors.Join(
			fmt.Errorf("access token error"),
			err,
		)
	}

	return rtoken, nil
}

func (ep *IdentityEndpoint[U]) getAccessTokenStringFromRequest(r router.Request) string {
	switch ep.Media {
	case PlainHeadersMedia:
		return r.Headers().Get(ep.accessTokenKey())
	default:
		if c := r.Cookie(ep.accessTokenKey()); c != nil {
			return c.Value
		}
	}

	return ""
}

func (ep *IdentityEndpoint[U]) getRefreshTokenStringFromRequest(r router.Request) string {
	switch ep.Media {
	case PlainHeadersMedia:
		return r.Headers().Get(ep.refreshTokenKey())
	default:
		if c := r.Cookie(ep.refreshTokenKey()); c != nil {
			return c.Value
		}
	}

	return ""
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

func (ep *IdentityEndpoint[U]) accessTokenKey() string {
	if len(ep.AccessTokenKey) == 0 {
		return DefaultAccessTokenKey
	}

	return ep.AccessTokenKey
}

func (ep *IdentityEndpoint[U]) refreshTokenKey() string {
	if len(ep.RefreshTokenKey) == 0 {
		return DefaultRefreshTokenKey
	}

	return ep.RefreshTokenKey
}

func (ep *IdentityEndpoint[U]) log() *slog.Logger {
	return log.OrDiscard(ep.Log)
}
