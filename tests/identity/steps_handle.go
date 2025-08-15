package identity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yandzee/go-svc/identity"
)

type StepsHandle struct {
	t *testing.T

	steps []*Step
}

type Step struct {
	Request         *http.Request
	ResponseCheckFn ResponseCheckFn
}

type ResponseCheckFn func(*testing.T, *httptest.ResponseRecorder)

func (sh *StepsHandle) CheckAuth(tokens identity.TokenPair, body any) *Step {
	step := &Step{
		Request:         makeRequest(sh.t, http.MethodGet, AuthCheckURL, tokens, body),
		ResponseCheckFn: nil,
	}

	sh.steps = append(sh.steps, step)
	return step
}

func (s *Step) CheckUnauthorized() {
	s.ResponseCheckFn = makeRespChecker(http.StatusUnauthorized)
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
