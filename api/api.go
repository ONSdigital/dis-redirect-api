package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// A list of error messages for Redirect API
var (
	ErrRedis = errors.New("redis returned an error")
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

	api.get("/redirects/{id}", api.getRedirect)
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
