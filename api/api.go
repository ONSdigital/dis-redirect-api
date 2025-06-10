package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

// RedirectAPI provides a struct to wrap the api around
type RedirectAPI struct {
	Router      *mux.Router
	RedisClient RedisClient
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, r *mux.Router, redisClient RedisClient) *RedirectAPI {
	api := &RedirectAPI{
		Router:      r,
		RedisClient: redisClient,
	}

	// TODO: remove hello world example handler route
	r.HandleFunc("/hello", HelloHandler(ctx)).Methods("GET")
	api.get("/redirects/{id}", api.getRedirect)
	return api
}

// get registers a GET http.HandlerFunc.
func (api *RedirectAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}
