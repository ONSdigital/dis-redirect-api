package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Redirect struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// UpsertRedirect handles the creation or update of redirects
func (api *API) UpsertRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	if !isValidBase64(id) {
		http.Error(w, "invalid base64 id", http.StatusBadRequest)
		return
	}

	fromDecoded, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		http.Error(w, "cannot decode id", http.StatusBadRequest)
		return
	}

	var redirect Redirect
	if err := json.NewDecoder(r.Body).Decode(&redirect); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if redirect.From != string(fromDecoded) {
		http.Error(w, "invalid request: 'from' field does not match base64-decoded 'id' in the URL", http.StatusBadRequest)
		return
	}

	if !isValidRelativePath(redirect.From) || !isValidRelativePath(redirect.To) {
		http.Error(w, "'from' and 'to' must be relative paths starting with '/'", http.StatusBadRequest)
		return
	}

	// Prevent redirect loops
	if redirect.From == redirect.To {
		http.Error(w, "'from' and 'to' cannot be the same", http.StatusBadRequest)
		return
	}

	existingValue, _ := api.Store.GetValue(ctx, redirect.From)

	err = api.Store.UpsertValue(ctx, redirect.From, redirect.To, 0)
	if err != nil {
		http.Error(w, "failed to save", http.StatusInternalServerError)
		return
	}

	// Return appropriate status code
	if existingValue == "" {
		w.WriteHeader(http.StatusCreated) // 201 Created — new key
	} else {
		w.WriteHeader(http.StatusOK) // 200 OK — overwritten
	}
}

// DeleteRedirect handles the deletion of a redirect
func (api *API) DeleteRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	if !isValidBase64(id) {
		http.Error(w, "invalid base64 id", http.StatusBadRequest)
		return
	}

	keyBytes, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		http.Error(w, "cannot decode id", http.StatusBadRequest)
		return
	}
	key := string(keyBytes)

	// Check if the key exists
	_, err = api.Store.GetValue(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "redirect not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to check redirect existence", http.StatusInternalServerError)
		return
	}

	// Proceed to delete
	if err := api.Store.DeleteValue(ctx, key); err != nil {
		http.Error(w, "failed to delete redirect", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidBase64(s string) bool {
	_, err := base64.URLEncoding.DecodeString(s)
	return err == nil
}

func isValidRelativePath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//")
}
