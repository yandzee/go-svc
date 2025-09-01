package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/yandzee/go-svc/identity"
	"github.com/yandzee/go-svc/router"
)

type GuardOptions struct {
	IsOptional          bool
	IsUserFetchDisabled bool
}

type GuardResult[U identity.User] struct {
	User        *U
	Tokens      identity.ValidatedTokenPair
	IsResponded bool
	Options     GuardOptions
}

func (ep *IdentityEndpoint[U]) Guard(
	rctx *router.RequestContext,
	opts ...GuardOptions,
) (GuardResult[U], error) {
	log := ep.log().With("method", "Guard")
	result := GuardResult[U]{}

	if len(opts) > 0 {
		result.Options = opts[0]
	}

	pair, err := ep.tokensFromRequest(rctx.Request)
	if err != nil {
		log.Error("tokensFromRequest failure", "err", err.Error())

		rctx.Response.String(
			http.StatusInternalServerError,
			"Auth check has failed: "+err.Error(),
		)

		result.IsResponded = true
		return result, err
	}

	result.Tokens = pair

	if !result.Options.IsOptional && !pair.HasValidAccess() {
		rctx.Response.String(
			http.StatusUnauthorized,
			"CurrentUser: access token is either invalid or absent",
		)

		result.IsResponded = true
		return result, nil
	}

	if result.Options.IsUserFetchDisabled || !pair.HasValidAccess() {
		return result, nil
	}

	usr, err := ep.Provider.GetTokenUser(rctx.Context(), pair.AccessToken.Token)
	if err != nil {
		log.Error("GetTokenUser failure", "err", err.Error())

		rctx.Response.String(
			http.StatusInternalServerError,
			"GetTokenUser: "+err.Error(),
		)

		result.IsResponded = true
		return result, err
	}

	result.User = usr
	return result, nil
}

func (gr *GuardResult[U]) IsAuthorized() bool {
	return gr.Tokens.HasValidAccess() && (gr.Options.IsOptional || gr.User != nil)
}

func (gr *GuardResult[U]) GetUserId() (uuid.UUID, bool) {
	if gr.User != nil {
		return (*gr.User).GetId(), true
	}

	return gr.Tokens.UserId()
}
