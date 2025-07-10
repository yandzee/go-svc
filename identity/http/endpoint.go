package http

import (
	"log/slog"
	"net/http"

	"github.com/yandzee/go-svc/identity"
	"github.com/yandzee/go-svc/log"
	"github.com/yandzee/go-svc/server/router"
)

type IdentityEndpoint[U any] struct {
	Provider identity.Provider[U]
	Log      *slog.Logger
}

func Wrap[U any](id identity.Provider[U]) *IdentityEndpoint[U] {
	return &IdentityEndpoint[U]{
		Provider: id,
	}
}

func (ep *IdentityEndpoint[U]) SignUp() router.Handler {
	log := ep.log()

	return func(w http.ResponseWriter, r *http.Request, ctx router.Context) {
		signupRequest := identity.SignupRequest{}
		jsoner := ctx.Jsoner()

		res := jsoner.DecodeRequest(w, r, &signupRequest)
		if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
			log.Error("Signup body parse failure", "err", msg)
			http.Error(w, msg, st)
			return
		}

		log.Debug("Signup", "request", signupRequest)
		if msg, ok := signupRequest.IsValid(); !ok {
			log.Debug("Signup invalid credentials", "msg", msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		signupResult, err := ep.Provider.SignUp(r.Context(), &signupRequest)
		if err != nil {
			log.Error("Signup failed", "err", err.Error())
			http.Error(w, "Signup failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		jsoner.EncodeResponse(w, signupResult)
	}
}

func (ep *IdentityEndpoint[U]) log() *slog.Logger {
	return log.OrDiscard(ep.Log)
}
