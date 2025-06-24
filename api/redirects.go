package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-redirect-api/apierrors"
	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/gorilla/mux"
)

// getRedirect gets the value of a key from redis
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]

	decodedKey, err := decodeBase64(redirectID)
	if err != nil {
		api.handleError(ctx, w, err, http.StatusBadRequest)
		return
	}

	redirect, err := api.Store.GetBundle(ctx, decodedKey)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("key %s not found", decodedKey)) {
			api.handleError(ctx, w, err, http.StatusNotFound)
		} else {
			api.handleError(ctx, w, apierrors.ErrRedis, http.StatusInternalServerError)
		}
		return
	}

	responseBody := models.Redirect{
		From: decodedKey,
		To:   redirect,
	}

	redirectResponse, err := json.Marshal(responseBody)
	if err != nil {
		api.handleError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectResponse); err != nil {
		api.handleError(ctx, w, err, http.StatusInternalServerError)
	}
}

// decodeBase64 returns the original string of a base64 encoded string
func decodeBase64(encodedKey string) (string, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return "", fmt.Errorf("key %s not base64", encodedKey)
	}

	return string(decodedKey), nil
}
