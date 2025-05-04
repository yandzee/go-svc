package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yandzee/go-svc/httputils"
	"github.com/yandzee/go-svc/server/router"
)

const (
	BaseURL         = "/test"
	ExtendBaseURL   = "/extended"
	AttachedBaseURL = "/attached"
)

type TestOutputs struct {
	// Mapping from route path to number of times handler was called
	Counter map[string]int
}

func TestMethods(t *testing.T) {
	r, outs := buildRouter(t)
	handler, err := r.Handler()

	if err != nil {
		t.Fatalf("Failed to build router.Handler(): %s\n", err.Error())
	}

	for _, method := range httputils.AllMethods {
		for _, baseUrl := range baseUrls() {
			path := baseUrl + "/" + strings.ToLower(method)
			req := httptest.NewRequest(method, path, nil)
			resp := httptest.NewRecorder()

			handler.ServeHTTP(resp, req)

			if num := outs.Counter[path]; num != 1 {
				t.Fatalf(
					"%s request to '%s' is handled wrong number of times: %d\n%v",
					method,
					path,
					num,
					outs,
				)
			}
		}
	}
}

func buildRouter(t *testing.T) (router.Router, *TestOutputs) {
	r := router.New()
	ext := router.New()
	att := router.New()

	outs := &TestOutputs{
		Counter: make(map[string]int),
	}

	for _, method := range httputils.AllMethods {
		methodStr := strings.ToLower(method)
		path := BaseURL + "/" + methodStr

		r.Method(method, path, func(
			w http.ResponseWriter,
			req *http.Request,
			ctx router.Context,
		) {
			outs.Counter[path] += 1
		})

		extPath := ExtendBaseURL + "/" + methodStr
		ext.Method(method, extPath, func(
			w http.ResponseWriter,
			req *http.Request,
			ctx router.Context,
		) {
			outs.Counter[extPath] += 1
		})

		att.Method(method, "/"+methodStr, func(
			w http.ResponseWriter,
			req *http.Request,
			ctx router.Context,
		) {
			path := AttachedBaseURL + "/" + methodStr
			outs.Counter[path] += 1
		})
	}

	if err := r.Extend(ext.IterRoutes()); err != nil {
		t.Fatalf("Failed to extend routes: %s", err.Error())
	}

	if err := r.Extend(att.IterRoutes(), AttachedBaseURL); err != nil {
		t.Fatalf("Failed to attach routes: %s", err.Error())
	}

	return r, outs
}

func baseUrls() []string {
	return []string{
		BaseURL,
		ExtendBaseURL,
		AttachedBaseURL,
	}
}
