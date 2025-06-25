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

// getRedirect gets the value of a key from the store
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]

	if !isValidBase64(redirectID) {
		http.Error(w, "invalid base64 id", http.StatusBadRequest)
		return
	}

	decodedString, err := base64.URLEncoding.DecodeString(redirectID)
	if err != nil {
		http.Error(w, "cannot decode id", http.StatusBadRequest)
		return
	}

	decodedKey := string(decodedString)

	redirect, err := api.Store.GetRedirect(ctx, decodedKey)
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

func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
