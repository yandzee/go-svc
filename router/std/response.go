package stdrouter

import (
	"fmt"
	"net/http"
)

type Response struct {
	Original http.ResponseWriter
}

func (r *Response) Write(d []byte) (int, error) {
	return r.Original.Write(d)
}

func (r *Response) Headers() http.Header {
	return r.Original.Header()
}

func (r *Response) String(code int, body ...string) {
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

func (r *Response) Stringf(code int, fmts string, args ...any) {
	switch {
	case code < 300:
		r.Original.WriteHeader(code)
		_, _ = fmt.Fprintf(r.Original, fmts, args...)
	default:
		http.Error(r.Original, fmt.Sprintf(fmts, args...), code)
	}
}
