package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/store"
	dpurl "github.com/ONSdigital/dis-redirect-api/url"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// RedirectAPI provides a struct to wrap the api around
type RedirectAPI struct {
	Router             *mux.Router
	RedirectStore      *store.Datastore
	authMiddleware     authorisation.Middleware
	urlBuilder         *dpurl.Builder
	enableURLRewriting bool
	apiUrl             *url.URL
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, r *mux.Router, dataStore *store.Datastore, auth authorisation.Middleware, cfg *config.Config, builder *dpurl.Builder) *RedirectAPI {
	apiURL, err := url.Parse(cfg.RedirectAPIURL)
	if err != nil {
		log.Error(ctx, "could not parse redirect api url", err, log.Data{"url": cfg.RedirectAPIURL})
		return nil
	}
	api := &RedirectAPI{
		Router:             r,
		RedirectStore:      dataStore,
		authMiddleware:     auth,
		urlBuilder:         builder,
		enableURLRewriting: cfg.EnableURLRewriting,
		apiUrl:             apiURL,
	}

	api.get("/v1/redirects/{id}", api.getRedirect)

	api.get("/v1/redirects", api.getRedirects)

	api.put(
		"/v1/redirects/{id}",
		auth.Require("legacy:edit", api.UpsertRedirect),
	)

	api.delete(
		"/v1/redirects/{id}",
		auth.Require("legacy:delete", api.DeleteRedirect),
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

func (api *RedirectAPI) delete(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodDelete)
}

// handleError returns the specified error message and HTTP code
func (api *RedirectAPI) handleError(ctx context.Context, w http.ResponseWriter, err error, status int) {
	log.Error(ctx, "request failed", err)
	http.Error(w, err.Error(), status)
}
