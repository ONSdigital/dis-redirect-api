package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// getRedirect gets the value of a key from redis
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]
	logData := log.Data{"redirect_id": redirectID}

	decodedKey, err := decodeBase64(redirectID)
	if err != nil {
		log.Error(ctx, "getRedirect endpoint: key not base64", err, logData)
		w.WriteHeader(http.StatusBadRequest)
	}

	redirect, err := api.RedisClient.GetValue(ctx, decodedKey)
	if err != nil {
		if errors.Is(err, fmt.Errorf("key %v not found", redirectID)) {
			log.Error(ctx, "getRedirect endpoint: key not found", err, logData)
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Error(ctx, "getRedirect endpoint: api.Redis.GetValue returned an error", err, logData)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	b, err := json.Marshal(redirect)
	if err != nil {
		log.Error(ctx, "error returned from json marshal", err, logData)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// decodeBase64 returns the original string of a base64 encoded string
func decodeBase64(encode string) (string, error) {
	decode, err := base64.StdEncoding.DecodeString(encode)

	if err != nil {
		return "", err
	}
	return string(decode), nil
}
