package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// createMD5Checksum returns the MD5 checksum of a url string
func createMD5Checksum(url string) string {
	urlHash := md5.Sum([]byte(url))
	return hex.EncodeToString(urlHash[:])
}

// getRedirect gets the value of a key from redis
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]
	logData := log.Data{"redirect_id": redirectID}
	hashedUrl := createMD5Checksum(redirectID)

	redirect, err := api.Redis.GetValue(ctx, hashedUrl)
	if err != nil {
		log.Error(ctx, "getRedirect endpoint: api.Redis.GetValue returned an error", err, logData)
		w.WriteHeader(http.StatusInternalServerError)
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
