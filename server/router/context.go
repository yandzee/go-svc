package router

import (
	"github.com/julienschmidt/httprouter"
	"github.com/yandzee/go-svc/httputils"
)

type HttprouterContext struct {
	ps     httprouter.Params
	jsoner *httputils.Jsoner
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

func (hc *HttprouterContext) Jsoner() *httputils.Jsoner {
	if hc.jsoner != nil {
		return hc.jsoner
	}

	hc.jsoner = &httputils.Jsoner{
		DefaultDecodeOptions: httputils.JSONDecodeOptions{
			MaxSize:              httputils.MaxSizeDefault,
			UnknownFieldsAllowed: false,
		},
	}

	return hc.jsoner
}
