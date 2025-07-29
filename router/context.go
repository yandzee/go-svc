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
	PathParam(string) (string, bool)
	LimitedBody(uint) io.ReadCloser
	URL() *url.URL
}

type Response interface {
	io.Writer
	StringResponder

	Headers() http.Header
}

type StringResponder interface {
	String(int, ...string)
	Stringf(int, string, ...any)
}

type RequestContext struct {
	Request  Request
	Response Response
}

func (rctx *RequestContext) Context() context.Context {
	return rctx.Request.Context()
}
