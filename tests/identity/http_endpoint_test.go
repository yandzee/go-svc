package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandzee/go-svc/identity"
	"github.com/yandzee/go-svc/identity/http"
)

const (
	AccessHeaderName  = "X-Test-Access-Token"
	RefreshHeaderName = "X-Test-Refresh-Token"
)

func buildEndpoint(t *testing.T) *http.IdentityEndpoint[TestUser] {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key for IdentityEndpoint: %s", err.Error())
	}

	inMemRegistry := &MockUserRegistry{}

	provider := identity.RegistryProvider[TestUser]{
		Registry:             inMemRegistry,
		BaseClaims:           jwt.RegisteredClaims{},
		TokenPrivateKey:      key,
		AccessTokenDuration:  0,
		RefreshTokenDuration: 0,
	}

	return &http.IdentityEndpoint[TestUser]{
		Provider:           &provider,
		Log:                nil,
		AccessTokenHeader:  AccessHeaderName,
		RefreshTokenHeader: RefreshHeaderName,
		TokenPrivateKey:    key,
	}
}

func TestEndpointBuild(t *testing.T) {
	_ = buildEndpoint(t)
}
