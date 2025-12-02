package stdrouter

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/yandzee/go-svc/data/jsoner"
	"github.com/yandzee/go-svc/router"
)

type Response struct {
	Original http.ResponseWriter
	Request  *http.Request
	Jsoner   *jsoner.Jsoner
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

func (r *Response) JSON(code int, d any, opts ...router.RespondOptions) (int, error) {
	hs := r.Original.Header()
	hs.Set("Content-Type", "application/json")

	buf := bytes.Buffer{}
	wr := bufio.NewWriter(&buf)

	err := r.Jsoner.Encode(wr, d)
	if err != nil {
		return 0, err
	}

	if err := wr.Flush(); err != nil {
		return 0, err
	}

	nbytes := buf.Len()
	hs.Set("Content-Length", strconv.Itoa(nbytes))

	if code != 0 {
		r.Original.WriteHeader(code)
	}

	_, err = r.Original.Write(buf.Bytes())
	return nbytes, err
}

func (r *Response) Redirect(code int, to string) {
	http.Redirect(r.Original, r.Request, to, code)
}

func (r *Response) SetCookie(c *http.Cookie) {
	http.SetCookie(r.Original, c)
}
