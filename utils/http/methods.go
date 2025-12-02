package httputils

import "net/http"

var (
	AllMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodHead,
		http.MethodOptions,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodPatch,
		http.MethodTrace,
	}
)
