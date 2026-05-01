package identity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yandzee/go-svc/identity"
	id_http "github.com/yandzee/go-svc/identity/http"
)

type StepsHandle struct {
	t *testing.T

	steps []*Step
}

type Step struct {
	Request         *http.Request
	ResponseChekers []ResponseCheckFn
}

type ResponseCheckFn func(*testing.T, *httptest.ResponseRecorder)

func (sh *StepsHandle) AddStep() *Step {
	s := &Step{}
	sh.steps = append(sh.steps, s)

	return s
}

func (s *Step) CheckAuth(tokens identity.TokenPair, body any) *Step {
	s.Request = s.createRequest(http.MethodGet, AuthCheckURL, tokens, body)
	return s
}

func (s *Step) Signin(creds identity.Credentials) *Step {
	s.Request = s.createRequest(
		http.MethodPost,
		SigninURL,
		identity.TokenPair{},
		identity.SigninRequest{
			Credentials: creds,
		},
	)

	return s
}

func (s *Step) Signup(creds identity.Credentials) *Step {
	s.Request = s.createRequest(
		http.MethodPost,
		SignupURL,
		identity.TokenPair{},
		identity.SignupRequest{
			Credentials: creds,
		},
	)

	return s
}

func (s *Step) ExpectStatus(statusCode int) {
	s.ResponseChekers = append(s.ResponseChekers, func(t *testing.T, rr *httptest.ResponseRecorder) {
		if rr.Code == statusCode {
			return
		}

		t.Fatalf("RespChecker: expected status %d, but got %d", statusCode, rr.Code)
	})
}

func (s *Step) ExpectTokens(media id_http.AuthCredentialsMedia, both bool) {
	s.ResponseChekers = append(s.ResponseChekers, func(t *testing.T, rr *httptest.ResponseRecorder) {
		switch media {
		case id_http.PlainHeadersMedia:
			h := rr.Result().Header

			at := h.Get(AccessTokenKey)
			rt := h.Get(RefreshTokenKey)

			if len(at) == 0 {
				t.Fatalf("response does not have AccessToken plain header, headers: %v", h)
			}

			if both && len(rt) == 0 {
				t.Fatalf("response does not have refresh token plain header, headers: %v", h)
			}

			if !both && len(rt) > 0 {
				t.Fatalf("response expected not to have refresh token attached, headers: %v", h)
			}
		}
	})
}

func (s *Step) createRequest(
	method, url string,
	tokens identity.TokenPair,
	body any,
) *http.Request {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		panic("makeRequest failed on body marshaling: " + err.Error())
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		panic("makeRequest failed on creaing new request: " + err.Error())
	}

	if tokens.AccessToken != nil {
		req.Header.Add(AccessTokenKey, tokens.AccessToken.JWTString)

		c := tokens.AccessToken.AsCookie(AccessTokenKey)
		req.AddCookie(&c)
	}

	if tokens.RefreshToken != nil {
		req.Header.Add(RefreshTokenKey, tokens.RefreshToken.JWTString)
	}

	return req
}

func makeRespChecker(status int) ResponseCheckFn {
	return func(t *testing.T, rr *httptest.ResponseRecorder) {
		if rr.Code != status {
			t.Fatalf("RespChecker: expected status %d, but got %d", status, rr.Code)
		}
	}
}
