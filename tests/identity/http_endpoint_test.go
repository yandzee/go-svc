package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandzee/go-svc/identity"
	"github.com/yandzee/go-svc/identity/http"
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
	_ = buildEndpoint(t, &TestDescriptor{
		AccessTokenDuration:  time.Minute,
		RefreshTokenDuration: time.Minute,
	})
}

func buildEndpoint(t *testing.T, td *TestDescriptor) *http.IdentityEndpoint[TestUser] {
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

	return &http.IdentityEndpoint[TestUser]{
		Provider:           &provider,
		Log:                nil,
		AccessTokenHeader:  AccessHeaderName,
		RefreshTokenHeader: RefreshHeaderName,
		TokenPrivateKey:    key,
	}
}
