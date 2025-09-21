package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandzee/go-svc/identity"
	id_http "github.com/yandzee/go-svc/identity/http"
	"github.com/yandzee/go-svc/router"
	stdrouter "github.com/yandzee/go-svc/router/std"
)

const (
	AccessHeaderName  = "X-Test-Access-Token"
	RefreshHeaderName = "X-Test-Refresh-Token"

	AuthCheckURL = "/auth"
	SigninURL    = "/auth/signin"
)

type TestDescriptor struct {
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Users                []TestUser
	Steps                func(h *StepsHandle)
}

func TestAuthCheckRoute(t *testing.T) {
	runTests(t, []TestDescriptor{
		{
			Steps: func(h *StepsHandle) {
				step := h.CheckAuth(identity.TokenPair{}, nil)
				step.ExpectStatus(http.StatusUnauthorized)

				step = h.CheckAuth(identity.TokenPair{
					AccessToken: &identity.Token{
						JWTString: "not-a-jwt",
					},
				}, nil)
				step.ExpectStatus(http.StatusUnauthorized)
			},
		},
	},
	)
}

func TestSigninRoute(t *testing.T) {
	runTests(t, []TestDescriptor{
		{
			Steps: func(h *StepsHandle) {
				step := h.Signin(nil)
				step.ExpectStatus(http.StatusBadRequest)
			},
		},
	})
}

func runTests(t *testing.T, tests []TestDescriptor) {
	for _, td := range tests {
		ep := buildEndpoint(t, &td)
		router := buildEndpointRouter(ep)

		stepsHandle := &StepsHandle{
			t: t,
		}

		td.Steps(stepsHandle)

		for _, step := range stepsHandle.steps {
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, step.Request)
			step.ResponseCheckFn(t, resp)
		}
	}
}

func buildEndpointRouter(ep *id_http.IdentityEndpoint[TestUser]) http.Handler {
	r := router.NewBuilder()

	r.Get(AuthCheckURL, ep.Check())
	r.Get("/auth/user", ep.CurrentUser())
	r.Post("/auth/signup", ep.Signup())
	r.Post(SigninURL, ep.Signin())
	r.Post("/auth/refresh", ep.Refresh())

	return stdrouter.Build(&r)
}

func buildEndpoint(t *testing.T, td *TestDescriptor) *id_http.IdentityEndpoint[TestUser] {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key for IdentityEndpoint: %s", err.Error())
	}

	inMemRegistry := &MockUserRegistry{
		Users: map[string]*TestUser{},
	}

	for _, usr := range td.Users {
		inMemRegistry.Users[usr.Username] = &usr
	}

	provider := identity.RegistryProvider[TestUser]{
		Registry:             inMemRegistry,
		BaseClaims:           jwt.RegisteredClaims{},
		TokenPrivateKey:      key,
		AccessTokenDuration:  td.AccessTokenDuration,
		RefreshTokenDuration: td.RefreshTokenDuration,
	}

	return &id_http.IdentityEndpoint[TestUser]{
		Provider:           &provider,
		Log:                nil,
		AccessTokenHeader:  AccessHeaderName,
		RefreshTokenHeader: RefreshHeaderName,
		TokenPrivateKey:    key,
	}
}
