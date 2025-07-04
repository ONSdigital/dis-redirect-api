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
	getRedirectBaseURL  = "http://localhost:29900/v1/redirects/"
	getRedirectsBaseURL = "http://localhost:29900/v1/redirects"
	existingBase64Key   = "L2Vjb25vbXkvb2xkLXBhdGg="
	validRedirect       = &models.Redirect{
		From: "/economy/old-path",
		To:   "/economy/new-path",
	}
	selfBaseURL      = "https://api.beta.ons.gov.uk/v1/redirects/"
	notANumber       = "this-is-not-a-number"
	economyBulletin1 = "/economy/mybulletin1"
	financeBulletin1 = "/finance/mybulletin1"
	economyBulletin2 = "/economy/mybulletin2"
	financeBulletin2 = "/finance/mybulletin2"
	economyBulletin3 = "/economy/mybulletin3"
	financeBulletin3 = "/finance/mybulletin3"
)

func GetRedirectAPIWithMocks(datastore store.Datastore) *api.RedirectAPI {
	ctx := context.Background()
	r := mux.NewRouter()

	return api.Setup(ctx, r, &datastore)
}

func TestGetRedirectEndpoint(t *testing.T) {
	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the id is valid and encoded in base64", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+existingBase64Key, http.NoBody)
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
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+nonBase64Key, http.NoBody)
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
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+nonExistentBase64Key, http.NoBody)
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
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+existingBase64Key, http.NoBody)
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

func TestGetRedirectsSuccessWithDefaultParams(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the count and cursor values are using the defaults", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			keyValuePairs := make(map[string]string)
			keyValuePairs[economyBulletin1] = financeBulletin1
			keyValuePairs[economyBulletin2] = financeBulletin2
			keyValuePairs[economyBulletin3] = financeBulletin3
			keyValuePairs["/economy/mybulletin4"] = "/finance/mybulletin4"
			keyValuePairs["/economy/mybulletin5"] = "/finance/mybulletin5"
			keyValuePairs["/economy/mybulletin6"] = "/finance/mybulletin6"
			keyValuePairs["/economy/mybulletin7"] = "/finance/mybulletin7"
			keyValuePairs["/economy/mybulletin8"] = "/finance/mybulletin8"
			keyValuePairs["/economy/mybulletin9"] = "/finance/mybulletin9"
			keyValuePairs["/economy/mybulletin10"] = "/finance/mybulletin10"

			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(ctx context.Context, matchPattern string, count int64, cursor uint64) (map[string]string, uint64, error) {
					return keyValuePairs, 0, nil
				},
				GetTotalKeysFunc: func(ctx context.Context) (int64, error) {
					return 12, nil
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirects
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				respRedirectList := response.RedirectList
				respItem1 := respRedirectList[0]
				respItem1From := respItem1.From
				expectedID := api.EncodeBase64(respItem1From)

				So(response.Count, ShouldEqual, 10)
				So(len(respRedirectList), ShouldEqual, 10)
				So(respItem1From, ShouldNotBeEmpty)
				So(respItem1.To, ShouldNotBeEmpty)
				So(respItem1.Id, ShouldEqual, expectedID)
				So(respItem1.Links.Self.Id, ShouldEqual, expectedID)
				So(respItem1.Links.Self.Href, ShouldEqual, selfBaseURL+expectedID)
				So(response.Cursor, ShouldEqual, "0")
				So(response.NextCursor, ShouldEqual, "0")
				So(response.TotalCount, ShouldEqual, 12)
			})
		})
	})
}

func TestGetRedirectsSuccessWithValidParams(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the count and cursor values are set to valid values", func() {
			countValue := "3"
			cursorValue := "1"
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?count="+countValue+"&cursor="+cursorValue, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			keyValuePairs := make(map[string]string)
			keyValuePairs[economyBulletin1] = financeBulletin1
			keyValuePairs[economyBulletin2] = financeBulletin2
			keyValuePairs[economyBulletin3] = financeBulletin3

			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(ctx context.Context, matchPattern string, count int64, cursor uint64) (map[string]string, uint64, error) {
					return keyValuePairs, 0, nil
				},
				GetTotalKeysFunc: func(ctx context.Context) (int64, error) {
					return 12, nil
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirects
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				respRedirectList := response.RedirectList
				respItem1 := respRedirectList[0]
				respItem1From := respItem1.From
				expectedID := api.EncodeBase64(respItem1From)

				So(response.Count, ShouldEqual, 3)
				So(len(respRedirectList), ShouldEqual, 3)
				So(respItem1From, ShouldNotBeEmpty)
				So(respItem1.To, ShouldNotBeEmpty)
				So(respItem1.Id, ShouldEqual, expectedID)
				So(respItem1.Links.Self.Id, ShouldEqual, expectedID)
				So(respItem1.Links.Self.Href, ShouldEqual, selfBaseURL+expectedID)
				So(response.Cursor, ShouldEqual, "1")
				So(response.NextCursor, ShouldEqual, "0")
				So(response.TotalCount, ShouldEqual, 12)
			})
		})
	})
}

func TestGetRedirectsCountNotAnInteger(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the count value given is not an integer", func() {
			countValue := notANumber
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?count="+countValue, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			mockStore := &storetest.StorerMock{}
			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetRedirectsCountNegative(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the count value given is negative", func() {
			countValue := "-12"
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?count="+countValue, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			mockStore := &storetest.StorerMock{}
			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetRedirectsCursorNotAnInteger(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the cursor value given is not an integer", func() {
			cursorValue := notANumber
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?cursor="+cursorValue, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			mockStore := &storetest.StorerMock{}
			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetRedirectsCursorNegative(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the cursor value given is negative", func() {
			cursorValue := "-7"
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?cursor="+cursorValue, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			mockStore := &storetest.StorerMock{}
			redirectAPI := GetRedirectAPIWithMocks(store.Datastore{Backend: mockStore})
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetRedirectsServerError(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the redirects server has an internal error", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(ctx context.Context, matchPattern string, count int64, cursor uint64) (map[string]string, uint64, error) {
					return nil, 0, apierrors.ErrRedis
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
