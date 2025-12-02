package router

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Request interface {
	Context() context.Context
	Headers() http.Header
	Cookie(string) *http.Cookie
	AllCookies() []*http.Cookie
	PathParam(string) (string, bool)
	LimitedBody(uint) io.ReadCloser
	URL() *url.URL
}

type Response interface {
	io.Writer
	StringResponder
	JSONResponder

	Headers() http.Header
	Redirect(int, string)
	SetCookie(*http.Cookie)
}

type StringResponder interface {
	String(int, ...string)
	Stringf(int, string, ...any)
}

type RespondOptions struct {
	Chunked bool
}

type JSONResponder interface {
	JSON(int, any, ...RespondOptions) (int, error)
}

type RequestContext struct {
	Request  Request
	Response Response
}

func (rctx *RequestContext) Context() context.Context {
	return rctx.Request.Context()
}
