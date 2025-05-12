package router

import "github.com/julienschmidt/httprouter"

type HttprouterContext struct {
	ps httprouter.Params
}

// NOTE: Gets path param from the route url, e. g. /route/:param
func (hc *HttprouterContext) Param(pname string) (string, bool) {
	for _, p := range hc.ps {
		if p.Key == pname {
			return p.Value, true
		}
	}

	return "", false
}
