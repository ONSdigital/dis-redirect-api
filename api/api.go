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
	Router *mux.Router
	Store  *store.Datastore
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, r *mux.Router, store *store.Datastore) *RedirectAPI {
	api := &RedirectAPI{
		Router: r,
		Store:  store,
	}

	api.get("/redirects/{id}", api.getRedirect)
	api.get("/redirects", api.getRedirects)
	return api
}

// get registers a GET http.HandlerFunc.
func (api *RedirectAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

// handleError returns the specified error message and HTTP code
func (api *RedirectAPI) handleError(ctx context.Context, w http.ResponseWriter, origErr, newErr error, status int, logMsg string, logData log.Data) {
	log.Error(ctx, logMsg, origErr, logData)
	http.Error(w, newErr.Error(), status)
}
