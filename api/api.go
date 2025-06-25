package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-redirect-api/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// RedirectAPI provides a struct to wrap the api around
type RedirectAPI struct {
	Router         *mux.Router
	Store          *store.Datastore
	authMiddleware authorisation.Middleware
}

// Setup function sets up the api and returns an api
func Setup(r *mux.Router, dataStore *store.Datastore, auth authorisation.Middleware) *API {
	api := &RedirectAPI{
		Router:         r,
		Store:          dataStore,
		authMiddleware: auth,
	}

	api.get("/v1/redirects/{id}", api.getRedirect)

	api.put(
		"/v1/redirects/{id}",
		auth.Require("legacy:edit", api.UpsertRedirect),
	)

	return api
}

// get registers a GET http.HandlerFunc.
func (api *RedirectAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

func (api *RedirectAPI) put(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodPut)
}

// handleError returns the specified error message and HTTP code
func (api *RedirectAPI) handleError(ctx context.Context, w http.ResponseWriter, err error, status int) {
	log.Error(ctx, "request failed", err)
	http.Error(w, err.Error(), status)
}
