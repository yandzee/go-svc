package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
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
)

type TestDescriptor struct {
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Users                []TestUser
}

func TestEndpointBuild(t *testing.T) {
	ep := buildEndpoint(t, &TestDescriptor{
		AccessTokenDuration:  time.Minute,
		RefreshTokenDuration: time.Minute,
	})

	_ = buildEndpointRouter(t, nil, ep)
}

func buildEndpointRouter(
	t *testing.T,
	td *TestDescriptor,
	ep *id_http.IdentityEndpoint[TestUser],
) http.Handler {
	r := router.NewBuilder()

	r.Get("/auth", ep.Check())
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
