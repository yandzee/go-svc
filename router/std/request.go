package stdrouter

import "net/http"

type Request struct {
	Original *http.Request
}

func (r *Request) PathParam(key string) (string, bool) {
	p := r.Original.PathValue(key)

	return p, len(p) > 0
}

func (r *Request) Headers() http.Header {
	return r.Original.Header
}
