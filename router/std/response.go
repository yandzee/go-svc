package stdrouter

import (
	"fmt"
	"net/http"
)

type Response struct {
	Original http.ResponseWriter
}

func (r *Response) Status(code int, body ...string) {
	switch {
	case code < 300:
		r.Original.WriteHeader(code)

		if len(body) > 0 {
			_, _ = fmt.Fprintln(r.Original, body[0])
		}
	default:
		t := ""
		if len(body) > 0 {
			t = body[0]
		}

		http.Error(r.Original, t, code)
	}
}

func (r *Response) Statusf(code int, fmts string, args ...any) {
	switch {
	case code < 300:
		r.Original.WriteHeader(code)
		_, _ = fmt.Fprintf(r.Original, fmts, args...)
	default:
		http.Error(r.Original, fmt.Sprintf(fmts, args...), code)
	}
}
