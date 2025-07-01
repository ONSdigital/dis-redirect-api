package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/log.go/v2/log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-redirect-api/apierrors"
	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// getRedirect gets the value of a key from the store
func (api *RedirectAPI) getRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	redirectID := vars["id"]

	decodedKey, err := decodeBase64(redirectID)
	if err != nil {
		logData := log.Data{"redirect_id": redirectID}
		api.handleError(ctx, w, err, err, http.StatusBadRequest, "request failed", logData)
		return
	}

	redirect, err := api.Store.GetRedirect(ctx, decodedKey)
	logData := log.Data{"redirect": redirect}
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("key %s not found", decodedKey)) {
			api.handleError(ctx, w, err, err, http.StatusNotFound, "request failed", nil)
		} else {
			api.handleError(ctx, w, err, apierrors.ErrRedis, http.StatusInternalServerError, "request failed", nil)
		}
		return
	}

	responseBody := models.Redirect{
		From: decodedKey,
		To:   redirect,
	}

	redirectResponse, err := json.Marshal(responseBody)
	if err != nil {
		api.handleError(ctx, w, err, err, http.StatusInternalServerError, "request failed", logData)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectResponse); err != nil {
		api.handleError(ctx, w, err, err, http.StatusInternalServerError, "request failed", logData)
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

// getRedirect gets the value of a key from the store
func (api *RedirectAPI) getRedirects(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	//vars := mux.Vars(req)
	//count := vars["count"]
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
	numRedirects, errCount := strconv.ParseInt(strCount, 10, 32)

	// validate count
	if errCount != nil {
		api.handleError(ctx, w, errCount, apierrors.ErrInvalidNumRedirects, http.StatusBadRequest, "invalid path parameter - failed to convert count to int64", logData)
		return
	}

	if numRedirects < 0 {
		errNegCount := errors.New("the count is negative")
		api.handleError(ctx, w, errNegCount, apierrors.ErrInvalidNumRedirects, http.StatusBadRequest, "invalid path parameter - count should be a positive integer", logData)
		return
	}

	// convert cursor from a string to a uint64 ready to use in call to GetRedirects
	var cursor uint64
	var errCursor error
	if strCursor != "0" {
		cursor, errCursor = strconv.ParseUint(strCursor, 10, 32)
	}

	logData = log.Data{"count": numRedirects, "cursor": cursor}

	// validate cursor
	if errCursor != nil {
		api.handleError(ctx, w, errCursor, apierrors.ErrInvalidCursor, http.StatusBadRequest, "invalid path parameter - failed to convert cursor to uint64", logData)
		return
	}

	keyValuePairs, newCursor, errRedirects := api.Store.GetRedirects(ctx, numRedirects, cursor)
	if errRedirects != nil {
		api.handleError(ctx, w, errRedirects, apierrors.ErrRedis, http.StatusInternalServerError, "error calling the store to get redirects", logData)
		return
	}

	var redirectList []models.Redirect

	for redirectPair := range keyValuePairs {
		var redirect models.Redirect
		redirectKey := redirectPair[0]
		key := strconv.Itoa(int(redirectKey))
		redirect.From = key
		redirectValue := redirectPair[1]
		value := strconv.Itoa(int(redirectValue))
		redirect.To = value
		redirectList = append(redirectList, redirect)
	}

	nextCursor := strconv.Itoa(int(newCursor))

	// To get the TotalCount we need to get the total number of redirects available in redis
	totalCount, errTotalCount := api.Store.GetTotalCount(ctx)
	if errTotalCount != nil {
		api.handleError(ctx, w, errTotalCount, apierrors.ErrRedis, http.StatusInternalServerError, "failed to get total count of redirects", logData)
		return
	}

	responseBody := models.Redirects{
		Count:        int(numRedirects),
		RedirectList: redirectList,
		Cursor:       strCursor,
		NextCursor:   nextCursor,
		TotalCount:   totalCount,
	}

	redirectsResponse, err := json.Marshal(responseBody)
	if err != nil {
		api.handleError(ctx, w, err, err, http.StatusInternalServerError, "failed to marshal response", logData)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(redirectsResponse); err != nil {
		api.handleError(ctx, w, err, err, http.StatusInternalServerError, "request failed", logData)
		return
	}

}
