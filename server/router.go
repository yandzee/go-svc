package server

import (
	"chelnok-backend/internal/data/page"
	"chelnok-backend/internal/server/handlers"

	"github.com/julienschmidt/httprouter"
)

func (srv *Server) prepareRouter() *httprouter.Router {
	router := httprouter.New()

	h := handlers.Handlers{
		Application: srv.Application,
		Log:         srv.Log.With("module", "server.Handlers"),
		Pager: &page.Pager{
			LimitKey:    "limit",
			OffsetKey:   "offset",
			LastSeenKey: "last-seen",
		},
	}

	// Authorization
	router.GET(srv.prefixed("/auth/sign"), h.GetSignMessage)
	router.GET(srv.prefixed("/auth"), h.CheckAuth)
	router.POST(srv.prefixed("/auth/signup"), h.Signup)
	router.POST(srv.prefixed("/auth/signin"), h.Signin)
	router.POST(srv.prefixed("/auth/signature"), h.PostSignature)
	router.POST(srv.prefixed("/auth/refresh"), h.RefreshAuth)

	// Board
	router.POST(srv.prefixed("/announce/moves"), h.GetAnnounceMovesPage)
	router.PUT(srv.prefixed("/announce/moves"), h.PutAnnounceMove)

	return router
}
