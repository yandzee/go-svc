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
		Request: sh.Request(http.MethodGet, AuthCheckURL, tokens, body),
	}

	sh.steps = append(sh.steps, step)
	return step
}

func (sh *StepsHandle) Signin(creds identity.Credentials) *Step {
	step := &Step{
		Request: sh.Request(http.MethodPost, SigninURL, identity.TokenPair{}, identity.SigninRequest{}),
	}

	sh.steps = append(sh.steps, step)
	return step
}

func (s *Step) ExpectStatus(statusCode int) {
	s.ResponseCheckFn = makeRespChecker(statusCode)
}

func makeRespChecker(status int) ResponseCheckFn {
	return func(t *testing.T, rr *httptest.ResponseRecorder) {
		if rr.Code != status {
			t.Fatalf("RespChecker: expected status %d, but got %d", status, rr.Code)
		}
	}
}

func (sh *StepsHandle) Request(
	method, url string,
	tokens identity.TokenPair,
	body any,
) *http.Request {
	t := sh.t

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
