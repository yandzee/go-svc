package router

import "github.com/julienschmidt/httprouter"

type HttprouterContext struct {
	ps httprouter.Params
}

func (hc *HttprouterContext) Param(pname string) (string, bool) {
	for _, p := range hc.ps {
		if p.Key == pname {
			return p.Value, true
		}
	}

	return "", false
}
