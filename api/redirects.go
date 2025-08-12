package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/ONSdigital/dp-net/v2/links"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// getRedirect gets the value of a key from the store
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]

	decodedString, err := base64.URLEncoding.DecodeString(redirectID)
	if err != nil {
		logData := log.Data{"redirect_id": redirectID}
		log.Info(ctx, "invalid base 64 id", logData)
		api.handleError(ctx, w, ErrInvalidBase64Id, http.StatusBadRequest)
		return
	}

	decodedKey := string(decodedString)

	redirect, err := api.RedirectStore.GetRedirect(ctx, decodedKey)
	logData := log.Data{"redirect": redirect}
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("key %s not found", decodedKey)) {
			log.Info(ctx, "redirect not found", logData)
			api.handleError(ctx, w, ErrNotFound, http.StatusNotFound)
		} else {
			log.Error(ctx, "redis failed on getting redirect", err, logData)
			api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		}
		return
	}

	responseBody := models.Redirect{
		From: decodedKey,
		To:   redirect,
	}

	redirectResponse, err := json.Marshal(responseBody)
	if err != nil {
		log.Error(ctx, "failed to marshal response", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectResponse); err != nil {
		log.Error(ctx, "failed to write response", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
	}
}

// UpsertRedirect handles the creation or update of redirects
func (api *RedirectAPI) UpsertRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	logData := log.Data{"redirect_id": id}

	fromDecoded, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		log.Info(ctx, "invalid base64 id", logData)
		api.handleError(ctx, w, ErrInvalidBase64Id, http.StatusBadRequest)
		return
	}

	var redirect models.Redirect
	if err := json.NewDecoder(r.Body).Decode(&redirect); err != nil {
		log.Info(ctx, "invalid redirect request")
		api.handleError(ctx, w, ErrInvalidRequestBody, http.StatusBadRequest)
		return
	}

	logData = log.Data{"redirect_from": redirect.From}
	if redirect.From != string(fromDecoded) {
		log.Info(ctx, "from field does not match base64 id", logData)
		api.handleError(ctx, w, ErrIDFromMismatch, http.StatusBadRequest)
		return
	}

	logData = log.Data{"redirect_from": redirect.From, "redirect_to": redirect.To}
	if !isValidRelativePath(redirect.From) || !isValidRelativePath(redirect.To) {
		log.Info(ctx, "from and to not relative paths", logData)
		api.handleError(ctx, w, ErrFromToNotRelative, http.StatusBadRequest)
		return
	}

	// Prevent redirect loops
	if redirect.From == redirect.To {
		log.Info(ctx, "'from' and 'to' cannot be the same", logData)
		api.handleError(ctx, w, ErrCircularPaths, http.StatusBadRequest)
		return
	}

	// Check if the redirect already exists but if not then create it
	existingValue, err := api.RedirectStore.GetValue(ctx, redirect.From)
	logData = log.Data{"existingValue": existingValue}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// log the error but then continue and create new redirect
			log.Info(ctx, "redirect not found so creating new one", logData)
		} else {
			log.Error(ctx, "redis failed on checking redirect existence", err, logData)
			api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
			return
		}
	}

	err = api.RedirectStore.UpsertValue(ctx, redirect.From, redirect.To, 0)
	if err != nil {
		log.Error(ctx, "redis failed on upserting redirect", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
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
	logData := log.Data{"redirect_id": id}

	keyBytes, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		log.Info(ctx, "invalid base 64 id", logData)
		api.handleError(ctx, w, ErrInvalidBase64Id, http.StatusBadRequest)
		return
	}
	key := string(keyBytes)

	// Check if the key exists
	logData = log.Data{"key": key}
	_, err = api.RedirectStore.GetValue(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Info(ctx, "redirect not found", logData)
			api.handleError(ctx, w, ErrNotFound, http.StatusNotFound)
			return
		}
		log.Error(ctx, "redis failed on checking redirect existence", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		return
	}

	// Proceed to delete
	if err := api.RedirectStore.DeleteValue(ctx, key); err != nil {
		log.Error(ctx, "redis failed on deleting redirect", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidRelativePath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//")
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
		log.Info(ctx, "invalid path parameter - failed to convert count to int64", logData)
		api.handleError(ctx, w, ErrInvalidCount, http.StatusBadRequest)
		return
	}

	// validate count
	if count < 0 {
		log.Info(ctx, "invalid path parameter - count should be a positive integer", logData)
		api.handleError(ctx, w, ErrNegativeCount, http.StatusBadRequest)
		return
	}

	// convert cursor from a string to a uint64 ready to use in call to GetRedirects
	cursor, err := strconv.ParseUint(strCursor, 10, 32)
	if err != nil {
		log.Info(ctx, "invalid path parameter - failed to convert cursor to uint64", logData)
		api.handleError(ctx, w, ErrInvalidOrNegativeCursor, http.StatusBadRequest)
		return
	}
	logData = log.Data{"count": count, "cursor": cursor}

	keyValuePairs, newCursor, err := api.RedirectStore.GetRedirects(ctx, count, cursor)
	if err != nil {
		log.Error(ctx, "redis failed on getting redirects", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		return
	}

	logData = log.Data{"num_redirects": len(keyValuePairs)}
	log.Info(ctx, "redirects retrieved from redis", logData)

	redirectList := make([]models.Redirect, 0, len(keyValuePairs))

	linkBuilder := links.FromHeadersOrDefault(&req.Header, api.apiURL)

	for key, value := range keyValuePairs {
		var redirect models.Redirect
		redirectID := encodeBase64(key)
		redirectHref, err := linkBuilder.BuildLink(fmt.Sprintf("/v1/redirects/%s", redirectID))
		if err != nil {
			log.Error(ctx, "redirect builder failed to build link", err, logData)
			api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
			return
		}
		redirectSelf := models.RedirectSelf{
			Href: redirectHref,
			ID:   redirectID,
		}
		redirectLinks := models.RedirectLinks{
			Self: redirectSelf,
		}
		redirect.From = key
		redirect.To = value
		redirect.ID = redirectID
		redirect.Links = redirectLinks
		redirectList = append(redirectList, redirect)
	}

	nextCursor := strconv.FormatUint(newCursor, 10)

	// To get the TotalCount we need to get the total number of redirects available in redis
	totalCount, errTotalCount := api.RedirectStore.GetTotalCount(ctx)
	if errTotalCount != nil {
		log.Error(ctx, "redis failed on getting total count of redirects", errTotalCount, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
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
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectsResponse); err != nil {
		log.Error(ctx, "failed to write response", err, logData)
		api.handleError(ctx, w, ErrInternal, http.StatusInternalServerError)
		return
	}
}
