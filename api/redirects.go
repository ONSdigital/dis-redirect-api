package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-redirect-api/apierrors"
	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// getRedirect gets the value of a key from the store
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]

	if !isValidBase64(redirectID) {
		errInvalidBase64 := errors.New("invalid base64 id")
		logData := log.Data{"redirect_id": redirectID}
		log.Error(ctx, "invalid base64 id", errInvalidBase64, logData)
		api.handleError(ctx, w, errInvalidBase64, http.StatusBadRequest)
		return
	}

	decodedString, err := base64.URLEncoding.DecodeString(redirectID)
	if err != nil {
		http.Error(w, "cannot decode id", http.StatusBadRequest)
		return
	}

	decodedKey := string(decodedString)

	redirect, err := api.RedirectStore.GetRedirect(ctx, decodedKey)
	logData := log.Data{"redirect": redirect}
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("key %s not found", decodedKey)) {
			api.handleError(ctx, w, err, http.StatusNotFound)
		} else {
			log.Error(ctx, "getting redirect from redis failed", err, logData)
			api.handleError(ctx, w, apierrors.ErrInternal, http.StatusInternalServerError)
		}
		return
	}

	responseBody := models.Redirect{
		From: decodedKey,
		To:   redirect,
	}

	redirectResponse, err := json.Marshal(responseBody)
	if err != nil {
		log.Error(ctx, "request failed", err, logData)
		api.handleError(ctx, w, apierrors.ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectResponse); err != nil {
		log.Error(ctx, "failed to write response", err, logData)
		api.handleError(ctx, w, apierrors.ErrInternal, http.StatusInternalServerError)
	}
}

// UpsertRedirect handles the creation or update of redirects
func (api *RedirectAPI) UpsertRedirect(w http.ResponseWriter, r *http.Request) {
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

	var redirect models.Redirect
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

	existingValue, _ := api.RedirectStore.GetValue(ctx, redirect.From)

	err = api.RedirectStore.UpsertValue(ctx, redirect.From, redirect.To, 0)
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
func (api *RedirectAPI) DeleteRedirect(w http.ResponseWriter, r *http.Request) {
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
	_, err = api.RedirectStore.GetValue(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "redirect not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to check redirect existence", http.StatusInternalServerError)
		return
	}

	// Proceed to delete
	if err := api.RedirectStore.DeleteValue(ctx, key); err != nil {
		http.Error(w, "failed to delete redirect", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidRelativePath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//")
}

func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// encodeBase64 returns the base64 encoded string of the original URL key string
func encodeBase64(key string) string {
	encodedKey := base64.StdEncoding.EncodeToString([]byte(key))

	return encodedKey
}

// getRedirects gets a paged list of redirects from the store
func (api *RedirectAPI) getRedirects(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	strCount := req.URL.Query().Get("count")
	strCursor := req.URL.Query().Get("cursor")

	// make count default to 10
	if strCount == "" {
		strCount = "10"
	}

	// make cursor default to 0
	if strCursor == "" {
		strCursor = "0"
	}

	logData := log.Data{"count": strCount, "cursor": strCursor}

	// convert count from a string to an int64 ready to use in call to GetRedirects
	count, errCount := strconv.ParseInt(strCount, 10, 32)

	if errCount != nil {
		log.Error(ctx, "invalid path parameter - failed to convert count to int64", errCount, logData)
		api.handleError(ctx, w, apierrors.ErrInvalidCount, http.StatusBadRequest)
		return
	}

	// validate count
	if count < 0 {
		errNegCount := errors.New("the count is negative")
		log.Error(ctx, "invalid path parameter - count should be a positive integer", errNegCount, logData)
		api.handleError(ctx, w, apierrors.ErrNegativeCount, http.StatusBadRequest)
		return
	}

	// convert cursor from a string to a uint64 ready to use in call to GetRedirects
	var cursor uint64
	var errCursor error
	cursor, errCursor = strconv.ParseUint(strCursor, 10, 32)

	logData = log.Data{"count": count, "cursor": cursor}

	// validate cursor
	if errCursor != nil {
		log.Error(ctx, "invalid path parameter - failed to convert cursor to uint64", errCursor, logData)
		api.handleError(ctx, w, apierrors.ErrInvalidOrNegativeCursor, http.StatusBadRequest)
		return
	}

	keyValuePairs, newCursor, errRedirects := api.RedirectStore.GetRedirects(ctx, count, cursor)
	if errRedirects != nil {
		log.Error(ctx, "error calling the store to get redirects", errRedirects, logData)
		api.handleError(ctx, w, apierrors.ErrInternal, http.StatusInternalServerError)
		return
	}

	logData = log.Data{"numKeyValuePairs": len(keyValuePairs)}
	log.Info(ctx, "The number of redirects fetched", logData)

	redirectBase := "https://api.beta.ons.gov.uk/v1/redirects/"
	redirectList := make([]models.Redirect, 0, len(keyValuePairs))

	for key, value := range keyValuePairs {
		var redirect models.Redirect
		redirectID := encodeBase64(key)
		redirectHref := redirectBase + redirectID
		redirectSelf := models.RedirectSelf{
			Href: redirectHref,
			Id:   redirectID,
		}
		redirectLinks := models.RedirectLinks{
			Self: redirectSelf,
		}
		redirect.From = key
		redirect.To = value
		redirect.Id = redirectID
		redirect.Links = redirectLinks
		redirectList = append(redirectList, redirect)
	}

	nextCursor := strconv.FormatUint(newCursor, 10)

	// To get the TotalCount we need to get the total number of redirects available in redis
	totalCount, errTotalCount := api.RedirectStore.GetTotalCount(ctx)
	if errTotalCount != nil {
		log.Error(ctx, "failed to get total count of redirects", errTotalCount, logData)
		api.handleError(ctx, w, apierrors.ErrInternal, http.StatusInternalServerError)
		return
	}

	responseBody := models.Redirects{
		Count:        int(count),
		RedirectList: redirectList,
		Cursor:       strCursor,
		NextCursor:   nextCursor,
		TotalCount:   totalCount,
	}

	redirectsResponse, err := json.Marshal(responseBody)
	if err != nil {
		log.Error(ctx, "failed to marshal response", err, logData)
		api.handleError(ctx, w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectsResponse); err != nil {
		log.Error(ctx, "request failed", err, logData)
		api.handleError(ctx, w, err, http.StatusInternalServerError)
		return
	}
}
