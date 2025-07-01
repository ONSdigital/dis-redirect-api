package api_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dis-redirect-api/api"
	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

const nonRedirectURL = "/non-redirect-url"
const testFromURL = "/foo"

func GetRedirectAPIWithMocks(datastore store.Datastore) *api.API {
	r := mux.NewRouter()

	return api.Setup(r, &datastore, newAuthMiddlwareMock())
}

func TestUpsertRedirect(t *testing.T) {
	Convey("Given a valid UpsertRedirect handler", t, func() {
		mockStore := &storetest.StorerMock{
			GetValueFunc: func(ctx context.Context, key string) (string, error) {
				switch key {
				case "/old-url":
					return "http://localhost:8081/new-url", nil
				case nonRedirectURL:
					return "", nil
				default:
					return "", nil
				}
			},
			SetValueFunc: func(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
				return nil
			},
		}

		apiInstance := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})

		Convey("When request is valid with matching base64 ID and from path", func() {
			from := testFromURL
			to := "/bar"
			id := base64.URLEncoding.EncodeToString([]byte(from))

			redirect := api.Redirect{
				From: from,
				To:   to,
			}
			body, _ := json.Marshal(redirect)
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusCreated)
			So(mockStore.SetValueCalls()[0].Key, ShouldEqual, from)
		})

		Convey("When ID is not valid base64", func() {
			req := httptest.NewRequest(http.MethodPut, "/redirects/!badid", bytes.NewBuffer([]byte(`{}`)))
			req = mux.SetURLVars(req, map[string]string{"id": "!badid"})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When base64 decodes but doesn't match body.from", func() {
			id := base64.URLEncoding.EncodeToString([]byte(testFromURL))
			redirect := api.Redirect{
				From: "/mismatch",
				To:   "/bar",
			}
			body, _ := json.Marshal(redirect)
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When Redis returns an error", func() {
			mockStore.SetValueFunc = func(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
				return errors.New("redis error")
			}

			from := testFromURL
			id := base64.URLEncoding.EncodeToString([]byte(from))
			redirect := api.Redirect{
				From: from,
				To:   "/bar",
			}
			body, _ := json.Marshal(redirect)
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{"id": id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When request body is invalid JSON", func() {
			id := base64.URLEncoding.EncodeToString([]byte("/foo"))
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer([]byte(`{bad json`)))
			req = mux.SetURLVars(req, map[string]string{"id": id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestDeleteRedirect(t *testing.T) {
	Convey("Given a DeleteRedirect handler", t, func() {
		mockStore := &storetest.StorerMock{}

		apiInstance := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})

		router := mux.NewRouter()
		router.HandleFunc("/redirects/{id}", apiInstance.DeleteRedirect).Methods(http.MethodDelete)

		// Helper to encode base64
		base64ID := base64.URLEncoding.EncodeToString([]byte("/test-path"))

		Convey("When the redirect exists and is deleted successfully", func() {
			mockStore.GetValueFunc = func(ctx context.Context, key string) (string, error) {
				if key == "/test-path" {
					return "/target", nil
				}
				return "", redis.Nil
			}
			mockStore.DeleteValueFunc = func(ctx context.Context, key string) error {
				return nil
			}

			req := httptest.NewRequest(http.MethodDelete, "/redirects/"+base64ID, http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusNoContent)
		})

		Convey("When the redirect does not exist", func() {
			mockStore.GetValueFunc = func(ctx context.Context, key string) (string, error) {
				return "", fmt.Errorf("redirect not found")
			}

			req := httptest.NewRequest(http.MethodDelete, "/redirects/"+base64ID, http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusNotFound)
			So(rr.Body.String(), ShouldContainSubstring, "redirect not found")
		})

		Convey("When an internal error occurs during existence check", func() {
			mockStore.GetValueFunc = func(ctx context.Context, key string) (string, error) {
				return "", errors.New("connection failed")
			}

			req := httptest.NewRequest(http.MethodDelete, "/redirects/"+base64ID, http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusInternalServerError)
			So(rr.Body.String(), ShouldContainSubstring, "failed to check redirect existence")
		})

		Convey("When the base64 id is invalid", func() {
			req := httptest.NewRequest(http.MethodDelete, "/redirects/invalid_base64", http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusBadRequest)
			So(rr.Body.String(), ShouldContainSubstring, "invalid base64 id")
		})
	})
}
