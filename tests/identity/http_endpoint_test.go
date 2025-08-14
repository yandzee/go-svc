package identity

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
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
)

type TestDescriptor struct {
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Users                []TestUser
	Steps                []StepFn
}

type ResponseCheckFn func(*testing.T, *httptest.ResponseRecorder)
type StepFn func() (*http.Request, ResponseCheckFn)

func TestAuthCheckRoute(t *testing.T) {
	runTests(t, []TestDescriptor{
		{
			Steps: []StepFn{
				StepFn(func() (*http.Request, ResponseCheckFn) {
					return makeRequest(
						t,
						http.MethodGet,
						AuthCheckURL,
						identity.TokenPair{},
						nil,
					), makeRespChecker(http.StatusUnauthorized)
				}),
				StepFn(func() (*http.Request, ResponseCheckFn) {
					return makeRequest(
						t,
						http.MethodGet,
						AuthCheckURL,
						identity.TokenPair{
							AccessToken: &identity.Token{
								JWTString: "not-a-jwt",
							},
						},
						nil,
					), makeRespChecker(http.StatusUnauthorized)
				}),
			},
		},
	})
}

func makeRespChecker(status int) ResponseCheckFn {
	return func(t *testing.T, rr *httptest.ResponseRecorder) {
		if rr.Code != status {
			t.Fatalf("RespChecker: expected status %d, but got %d", status, rr.Code)
		}
	}
}

func makeRequest(
	t *testing.T,
	method, url string,
	tokens identity.TokenPair,
	body any,
) *http.Request {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("makeRequest failed on body marshaling: %s", err.Error())
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("makeRequest failed on creaing new request: %s", err.Error())
	}

	if tokens.AccessToken != nil {
		req.Header.Add(AccessHeaderName, tokens.AccessToken.JWTString)

		c := tokens.AccessToken.AsCookie(AccessHeaderName)
		req.AddCookie(&c)
	}

	if tokens.RefreshToken != nil {
		req.Header.Add(RefreshHeaderName, tokens.RefreshToken.JWTString)
	}

	return req
}

func runTests(t *testing.T, tests []TestDescriptor) {
	for _, td := range tests {
		ep := buildEndpoint(t, &td)
		router := buildEndpointRouter(ep)

		for _, stepFn := range td.Steps {
			resp := httptest.NewRecorder()
			req, respChecker := stepFn()

			router.ServeHTTP(resp, req)
			respChecker(t, resp)
		}
	}
}

func buildEndpointRouter(ep *id_http.IdentityEndpoint[TestUser]) http.Handler {
	r := router.NewBuilder()

	r.Get(AuthCheckURL, ep.Check())
	r.Get("/auth/user", ep.CurrentUser())
	r.Post("/auth/signup", ep.Signup())
	r.Post("/auth/signin", ep.Signin())
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
