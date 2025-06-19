package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/api/apimock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	baseURL           = "http://localhost:29900/redirects/"
	existingBase64Key = "ZWNvbm9teS9vbGQtcGF0aA=="
)

func TestGetRedirectEndpoint(t *testing.T) {
	Convey("Given an healthy Redirect handler", t, func() {
		redisClientMock := &apimock.RedisClientMock{
			GetValueFunc: func(ctx context.Context, key string) (string, error) {
				return "new-path", nil
			},
		}

		redirectAPI := setupAPI(redisClientMock)

		Convey("When a redirect is requested with a valid key encoded in base64", func() {
			request := httptest.NewRequest(http.MethodGet, baseURL+existingBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("The matched redirect is found with status code 200", func() {
				payload, _ := io.ReadAll(responseRecorder.Body)
				_, err := json.Marshal(payload)
				So(err, ShouldBeNil)
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestGetRedirectReturns400(t *testing.T) {
	Convey("Given an healthy Redirect handler", t, func() {
		redisClientMock := &apimock.RedisClientMock{
			GetValueFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("getRedirect endpoint: key not base64")
			},
		}

		redirectAPI := setupAPI(redisClientMock)

		Convey("When a redirect is requested with a non base64 key", func() {
			var nonBase64Key = "some-string"
			request := httptest.NewRequest(http.MethodGet, baseURL+nonBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("The unmatched redirect is not found with status code 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetRedirectReturns404(t *testing.T) {
	Convey("Given an healthy Redirect handler", t, func() {
		redisClientMock := &apimock.RedisClientMock{
			GetValueFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("key old-path not found")
			},
		}

		redirectAPI := setupAPI(redisClientMock)

		Convey("When a redirect is requested with a non-existent key encoded in base64", func() {
			var nonExistentBase64Key = "b2xkLXBhdGg="
			request := httptest.NewRequest(http.MethodGet, baseURL+nonExistentBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("The unmatched redirect is not found with status code 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

func TestGetRedirectReturns500(t *testing.T) {
	Convey("Given a failing Redirect handler", t, func() {
		redisClientMock := &apimock.RedisClientMock{
			GetValueFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("redis returned an error")
			},
		}

		redirectAPI := setupAPI(redisClientMock)

		Convey("When a redirect is requested with a valid key encoded in base64", func() {
			request := httptest.NewRequest(http.MethodGet, baseURL+existingBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("The matched redirect is not found with status code 500", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
