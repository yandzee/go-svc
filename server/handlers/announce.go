package handlers

import (
	"chelnok-backend/internal/data/announce"
	"chelnok-backend/internal/data/board"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *Handlers) GetAnnounceMovesPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	pageSelector, err := h.Pager.FromHTTPRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	svcs := h.Application.Services()

	auth, err := svcs.Auth().FromHTTPRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filter := &announce.MoveFilter{}
	res := h.ensureJsoner().DecodeRequest(w, r, &filter)
	if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
		h.Log.Error("GetAnnounceMovesPage: failed to decode filter", "err", msg)
		http.Error(w, msg, st)
		return
	}

	h.Log.Debug("MOVE FILTERS", "f", filter)
	data, err := svcs.Board().GetMoveCards(r.Context(), auth, pageSelector, filter)
	if err != nil {
		h.Log.Error("GetMoves failed", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.Jsoner.EncodeResponse(w, data); err != nil {
		h.Log.Error("Failed to send moves page", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func (h *Handlers) PutAnnounceMove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	svcs := h.Application.Services()

	auth, err := svcs.Auth().FromHTTPRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log.Debug("PutAnnounceMove tokens", "tokens", auth)
	if !auth.HasValidAccess() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	m := board.MoveCard{}

	res := h.ensureJsoner().DecodeRequest(w, r, &m)
	if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
		h.Log.Error("PutAnnounceMove body parse failure", "err", msg)
		http.Error(w, msg, st)
		return
	}

	h.Log.Info("About to put move", "move", m)
	upserted, err := svcs.Board().PutMoveCard(r.Context(), auth, &m)
	if err != nil {
		h.Log.Error("svc.PutMove failed", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log.Debug("After PutMove", "upserted", upserted)

	if err := h.Jsoner.EncodeResponse(w, upserted); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
