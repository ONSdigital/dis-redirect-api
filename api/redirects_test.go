package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
)

func GetRedirectAPIWithMocks(datastore store.Datastore) *api.RedirectAPI {
	r := mux.NewRouter()

	return api.Setup(r, &datastore)
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
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
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
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
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
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
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
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
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
