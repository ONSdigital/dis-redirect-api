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
	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
	"github.com/gorilla/mux"
	"github.com/smartystreets/goconvey/convey"
)

const nonRedirectURL = "/non-redirect-url"
const testFromURL = "/foo"

func GetRedirectAPIWithMocks(datastore store.Datastore) *api.API {
	r := mux.NewRouter()

	return api.Setup(r, &datastore, newAuthMiddlwareMock())
}

func TestUpsertRedirect(t *testing.T) {
	convey.Convey("Given a valid UpsertRedirect handler", t, func() {
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

		convey.Convey("When request is valid with matching base64 ID and from path", func() {
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

			convey.So(rec.Result().StatusCode, convey.ShouldEqual, http.StatusCreated)
			convey.So(mockStore.SetValueCalls()[0].Key, convey.ShouldEqual, from)
		})

		convey.Convey("When ID is not valid base64", func() {
			req := httptest.NewRequest(http.MethodPut, "/redirects/!badid", bytes.NewBuffer([]byte(`{}`)))
			req = mux.SetURLVars(req, map[string]string{"id": "!badid"})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			convey.So(rec.Result().StatusCode, convey.ShouldEqual, http.StatusBadRequest)
		})

		convey.Convey("When base64 decodes but doesn't match body.from", func() {
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

			convey.So(rec.Result().StatusCode, convey.ShouldEqual, http.StatusBadRequest)
		})

		convey.Convey("When Redis returns an error", func() {
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

			convey.So(rec.Result().StatusCode, convey.ShouldEqual, http.StatusInternalServerError)
		})

		convey.Convey("When request body is invalid JSON", func() {
			id := base64.URLEncoding.EncodeToString([]byte("/foo"))
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer([]byte(`{bad json`)))
			req = mux.SetURLVars(req, map[string]string{"id": id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			convey.So(rec.Result().StatusCode, convey.ShouldEqual, http.StatusBadRequest)
		})
	})
}
