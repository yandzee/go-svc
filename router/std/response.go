package stdrouter

import "net/http"

type Response struct {
	Original http.ResponseWriter
}
