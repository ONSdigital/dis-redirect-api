package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-redirect-api/store"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// RedirectAPI provides a struct to wrap the api around
type RedirectAPI struct {
	Router        *mux.Router
	RedirectStore *store.Datastore
}

// Setup function sets up the api and returns an api
func Setup(r *mux.Router, redirectStore *store.Datastore) *RedirectAPI {
	api := &RedirectAPI{
		Router:        r,
		RedirectStore: redirectStore,
	}

	api.get("/v1/redirects/{id}", api.getRedirect)
	api.get("/v1/redirects", api.getRedirects)
	return api
}

// get registers a GET http.HandlerFunc.
func (api *RedirectAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

// handleError returns the specified error message and HTTP code
func (api *RedirectAPI) handleError(ctx context.Context, w http.ResponseWriter, err error, status int) {
	log.Error(ctx, "request failed", err)
	http.Error(w, err.Error(), status)
}
