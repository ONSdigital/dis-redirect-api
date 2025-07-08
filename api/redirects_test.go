package api_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dis-redirect-api/api"
	"github.com/ONSdigital/dis-redirect-api/apierrors"
	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	baseURL           = "http://localhost:29900/v1/redirects/"
	existingBase64Key = "L2Vjb25vbXkvb2xkLXBhdGg="
	nonRedirectURL    = "/non-redirect-url"
	testFromURL       = "/foo"
)

func GetRedirectAPIWithMocks(datastore store.Datastore) *api.RedirectAPI {
	r := mux.NewRouter()

	return api.Setup(r, &datastore, newAuthMiddlwareMock())
}
func TestGetRedirectEndpoint(t *testing.T) {
	validRedirect := &models.Redirect{
		From: "/economy/old-path",
		To:   "/economy/new-path",
	}

	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the id is valid and encoded in base64", func() {
			request := httptest.NewRequest(http.MethodGet, baseURL+existingBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetValueFunc: func(ctx context.Context, key string) (string, error) {
					return "/economy/new-path", nil
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirect
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				So(response.From, ShouldEqual, validRedirect.From)
				So(response.To, ShouldEqual, validRedirect.To)
			})
		})
	})
}

func TestGetRedirectReturns400(t *testing.T) {
	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the id is not endcoded in base64", func() {
			var nonBase64Key = "some-string"
			request := httptest.NewRequest(http.MethodGet, baseURL+nonBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetValueFunc: func(ctx context.Context, key string) (string, error) {
					return "", errors.New("key some-string not base64")
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetRedirectReturns404(t *testing.T) {
	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the id is valid and encoded in base64", func() {
			var nonExistentBase64Key = "b2xkLXBhdGg="
			request := httptest.NewRequest(http.MethodGet, baseURL+nonExistentBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetValueFunc: func(ctx context.Context, key string) (string, error) {
					return "", errors.New("key old-path not found")
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 404", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

func TestGetRedirectReturns500(t *testing.T) {
	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the redirect handler fails", func() {
			request := httptest.NewRequest(http.MethodGet, baseURL+existingBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetValueFunc: func(ctx context.Context, key string) (string, error) {
					return "", apierrors.ErrRedis
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 500", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
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

			redirect := models.Redirect{
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
			redirect := models.Redirect{
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
			redirect := models.Redirect{
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
