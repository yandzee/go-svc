package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"testing/fstest"

	"github.com/yandzee/go-svc/httputils"
	"github.com/yandzee/go-svc/router"
	stdrouter "github.com/yandzee/go-svc/router/std"
)

const (
	BaseURL         = "/test"
	ExtendBaseURL   = "/extended"
	AttachedBaseURL = "/attached"
	FilesURL        = "/files/"

	TestFilename1    = "testfile1.dat"
	TestFilename2    = "testfile2.dat"
	TestFileContent1 = "test file content 1"
	TestFileContent2 = "test file content 2"
)

type TestOutputs struct {
	// Mapping from route path to number of times handler was called
	Counter map[string]int
}

func TestMethods(t *testing.T) {
	handler, outs := buildRouter(t)

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

func TestFiles(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, FilesURL+TestFilename1, nil)
	req2 := httptest.NewRequest(http.MethodGet, AttachedBaseURL+FilesURL+TestFilename2, nil)
	expectedContent := []string{TestFileContent1, TestFileContent2}

	for i, req := range []*http.Request{req1, req2} {
		handler, _ := buildRouter(t)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("wrong response code %d for file '%s'", resp.Code, req.URL.Path)
			return
		}

		if s := resp.Body.String(); s != expectedContent[i] {
			t.Fatalf("wrong response '%s' for file '%s'", s, req.URL.Path)
			return
		}
	}
}

func buildRouter(t *testing.T) (http.Handler, *TestOutputs) {
	r := router.NewBuilder()
	ext := router.NewBuilder()
	att := router.NewBuilder()

	outs := &TestOutputs{
		Counter: make(map[string]int),
	}

	for _, method := range httputils.AllMethods {
		methodStr := strings.ToLower(method)
		path := BaseURL + "/" + methodStr

		r.Method(method, path, func(rctx *router.RequestContext) {
			outs.Counter[path] += 1
		})

		extPath := ExtendBaseURL + "/" + methodStr
		ext.Method(method, extPath, func(rctx *router.RequestContext) {
			outs.Counter[extPath] += 1
		})

		att.Method(method, "/"+methodStr, func(ctx *router.RequestContext) {
			path := AttachedBaseURL + "/" + methodStr
			outs.Counter[path] += 1
		})
	}

	r.Files(FilesURL, fstest.MapFS{
		TestFilename1: {
			Data: []byte(TestFileContent1),
		},
	})

	att.Files(FilesURL, fstest.MapFS{
		TestFilename2: {
			Data: []byte(TestFileContent2),
		},
	})

	if err := r.Extend(ext.IterRoutes()); err != nil {
		t.Fatalf("Failed to extend routes: %s", err.Error())
	}

	if err := r.Extend(att.IterRoutes(), AttachedBaseURL); err != nil {
		t.Fatalf("Failed to attach routes: %s", err.Error())
	}

	return stdrouter.Build(&r), outs
}

func baseUrls() []string {
	return []string{
		BaseURL,
		ExtendBaseURL,
		AttachedBaseURL,
	}
}
