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
	"github.com/google/uuid"
	"github.com/yandzee/go-svc/identity"
	id_http "github.com/yandzee/go-svc/identity/http"
	"github.com/yandzee/go-svc/router"
	stdrouter "github.com/yandzee/go-svc/router/std"
)

const (
	AccessTokenKey  = "X-Test-Access-Token"
	RefreshTokenKey = "X-Test-Refresh-Token"

	AuthCheckURL = "/auth"
	SigninURL    = "/auth/signin"
	SignupURL    = "/auth/signup"

	// NOTE: This user exists in mock registry, Username2 does not
	Username1         = "username-1"
	Username1Password = "password-1"

	Username2         = "username-2"
	Username2Password = "password-2"
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
				step := h.AddStep().CheckAuth(identity.TokenPair{}, nil)
				step.ExpectStatus(http.StatusUnauthorized)

				step = h.AddStep().CheckAuth(identity.TokenPair{
					AccessToken: &identity.Token{
						JWTString: "not-a-jwt",
					},
				}, nil)
				step.ExpectStatus(http.StatusUnauthorized)
			},
		},
	})
}

func TestSigninRoute(t *testing.T) {
	runTests(t, []TestDescriptor{
		{
			Steps: func(h *StepsHandle) {
				step := h.AddStep().Signin(nil)
				step.ExpectStatus(http.StatusBadRequest)
			},
		},
	})
}

func TestSignupPlainHeaders(t *testing.T) {
	runTests(t, []TestDescriptor{
		{
			Steps: func(h *StepsHandle) {
				step := h.AddStep().Signup(identity.Credentials{
					"username": Username2,
					"password": Username2Password,
				})
				step.ExpectTokens(id_http.PlainHeadersMedia, true)
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

			for _, responseChecker := range step.ResponseChekers {
				responseChecker(t, resp)
			}
		}
	}
}

func buildEndpointRouter(ep *id_http.IdentityEndpoint[TestUser]) http.Handler {
	r := router.NewBuilder()

	r.Get(AuthCheckURL, ep.Check())
	r.Get("/auth/user", ep.CurrentUser())
	r.Post(SignupURL, ep.Signup())
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
		Users: map[string]*TestUser{
			Username1: {
				Id:       uuid.Must(uuid.NewUUID()),
				Username: Username1,
				Password: Username1Password,
			},
		},
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
		Provider:        &provider,
		Log:             nil,
		TokenPrivateKey: key,
		Media:           id_http.PlainHeadersMedia,
		AccessTokenKey:  AccessTokenKey,
		RefreshTokenKey: RefreshTokenKey,
	}
}
