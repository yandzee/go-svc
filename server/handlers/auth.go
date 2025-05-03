package handlers

import (
	"chelnok-backend/internal/application/core"
	"chelnok-backend/internal/data/auth"
	"chelnok-backend/internal/data/page"
	svcauth "chelnok-backend/internal/services/auth"
	"chelnok-backend/pkg/httputils"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Handlers struct {
	Log         *slog.Logger
	Application core.Core
	Jsoner      *httputils.Jsoner
	Pager       *page.Pager
}

func (h *Handlers) GetSignMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wallet := strings.TrimSpace(r.URL.Query().Get("wallet"))

	if len(wallet) == 0 {
		http.Error(w, "Empty wallet address", http.StatusBadRequest)
		return
	}

	msg, err := h.Application.Services().Auth().GetSignMessage(r.Context(), wallet)
	if err != nil {
		h.Log.Error("GetSignMessage failed", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log.Debug("in GetAuthNonce", "wallet", wallet, "msg", msg)

	w.WriteHeader(200)
	fmt.Fprint(w, msg)
}

func (h *Handlers) PostSignature(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	signature := auth.Signature{}

	res := h.ensureJsoner().DecodeRequest(w, r, &signature)
	if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
		h.Log.Error("PostSignature body parse failure", "err", msg)
		http.Error(w, msg, st)
		return
	}

	h.Log.Debug("PostSignature signature", "signature", signature)

	svcs := h.Application.Services()
	verif, err := svcs.Auth().VerifySignature(r.Context(), &signature)

	if err != nil {
		h.Log.Error("VerifySignature failure", "err", err.Error())
		http.Error(
			w,
			"Signature verification has failed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	// svcs.Auth().RespondWithVerification(w, verif)
	h.ensureJsoner().EncodeResponse(w, verif.ForClient())
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	creds := auth.SignupCredentials{}
	jsoner := h.ensureJsoner()

	res := jsoner.DecodeRequest(w, r, &creds)
	if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
		h.Log.Error("Signup body parse failure", "err", msg)
		http.Error(w, msg, st)
		return
	}

	h.Log.Debug("Signup", "credentials", creds)
	if msg, ok := creds.IsValid(); !ok {
		h.Log.Debug("Signup invalid credentials", "msg", msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	svcs := h.Application.Services()
	signupResult, err := svcs.Auth().Signup(r.Context(), &creds)
	if err != nil {
		h.Log.Error("Signup failed", "err", err.Error())
		http.Error(w, "Signup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsoner.EncodeResponse(w, signupResult)
}

func (h *Handlers) Signin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	creds := auth.SignupCredentials{}
	jsoner := h.ensureJsoner()

	res := jsoner.DecodeRequest(w, r, &creds)
	if st, msg := res.AsHTTPStatus(); st != http.StatusOK {
		h.Log.Error("Signin body parse failure", "err", msg)
		http.Error(w, msg, st)
		return
	}

	h.Log.Debug("Signin", "credentials", creds)
	if msg, ok := creds.IsValid(); !ok {
		h.Log.Debug("Signin invalid credentials", "msg", msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	svcs := h.Application.Services()
	signinResult, err := svcs.Auth().Signin(r.Context(), &creds)
	if err != nil {
		h.Log.Error("Signin failed", "err", err.Error())
		http.Error(w, "Signin failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsoner.EncodeResponse(w, signinResult)

	if signinResult.IsInvalid {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (h *Handlers) CheckAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	svcs := h.Application.Services()

	pair, err := svcs.Auth().FromHTTPRequest(r)
	if err != nil {
		h.Log.Error("CheckAuth failure", "err", err.Error())
		http.Error(
			w,
			"CheckAuth has failed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	switch {
	case pair.AccessToken == nil:
		http.Error(w, "Unauthorized: no access token", http.StatusUnauthorized)
	case pair.AccessToken.Validation.IsExpired:
		http.Error(w, "Unauthorized: token is expired", http.StatusUnauthorized)
	case pair.AccessToken.Validation.IsMalformed:
		http.Error(w, "Unauthorized: token is malformed", http.StatusUnauthorized)
	case pair.AccessToken.Validation.IsParseError:
		err := pair.AccessToken.Validation.Error
		http.Error(w, "CheckAuth: token parse error: "+err.Error(), http.StatusInternalServerError)
	case pair.AccessToken.Validation.Error != nil:
		err := pair.AccessToken.Validation.Error
		http.Error(w, "CheckAuth: unexpected error: "+err.Error(), http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) RefreshAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	svcs := h.Application.Services()

	pair, err := svcs.Auth().FromHTTPRequest(r)
	if err != nil {
		h.Log.Error("RefreshAuth failure", "err", err.Error())
		http.Error(
			w,
			"RefreshAuth has failed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	switch {
	case pair.RefreshToken == nil:
		http.Error(w, "Bad request: refresh token must be attached", http.StatusBadRequest)
	case pair.RefreshToken.Validation.IsExpired:
		http.Error(w, "Unauthorized: token is expired", http.StatusUnauthorized)
	case pair.RefreshToken.Validation.IsMalformed:
		http.Error(w, "Unauthorized: token is malformed", http.StatusUnauthorized)
	case pair.RefreshToken.Validation.IsParseError:
		err := pair.RefreshToken.Validation.Error
		http.Error(w, "RefreshAuth: token parse error"+err.Error(), http.StatusInternalServerError)
	case pair.RefreshToken.Validation.Error != nil:
		err := pair.RefreshToken.Validation.Error
		http.Error(w, "CheckAuth: unexpected error: "+err.Error(), http.StatusInternalServerError)
	}

	if pair.RefreshToken.Validation.Error != nil {
		return
	}

	// wallet, isPresented := pair.RefreshToken.Token.GetWalletAddress()
	// h.Log.Debug("Wallet from token", "wallet", wallet, "isPresented", isPresented)
	//
	// if !isPresented {
	// 	http.Error(
	// 		w,
	// 		"RefreshAuth: failed to extract wallet address from token",
	// 		http.StatusUnauthorized,
	// 	)
	// 	return
	// }

	tokenPair, err := svcs.Auth().RefreshTokens(pair.RefreshToken.Token)

	switch {
	case err == nil:
		svcs.Auth().RespondWithTokenPair(w, &tokenPair)
	case errors.Is(err, svcauth.ErrInvalidSubject):
		http.Error(w, "RefreshAuth: "+err.Error(), http.StatusUnauthorized)
	default:
		http.Error(
			w,
			"RefreshAuth: unexpected error: "+err.Error(),
			http.StatusInternalServerError,
		)
	}
}

func (h *Handlers) ensureJsoner() *httputils.Jsoner {
	if h.Jsoner != nil {
		return h.Jsoner
	}

	h.Jsoner = &httputils.Jsoner{}
	return h.Jsoner
}
