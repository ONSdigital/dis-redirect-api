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
	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/models"
	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
	disRedis "github.com/ONSdigital/dis-redis"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	getRedirectBaseURL  = "http://localhost:29900/v1/redirects/"
	getRedirectsBaseURL = "http://localhost:29900/v1/redirects"
	selfBaseURL         = "http://localhost:29900/redirects/" // TODO Change this to be "http://localhost:29900/v1/redirects/" when dp-net has been fixed
	existingBase64Key   = "L2Vjb25vbXkvb2xkLXBhdGg="
	// fwdPrefix mirrors the unexported forward lookup prefix in the store package
	fwdPrefix     = "fwd:"
	redirectFrom  = "/economy/old-path"
	redirectTo    = "/economy/new-path"
	validRedirect = &models.Redirect{
		From: redirectFrom,
		To:   redirectTo,
	}
	notANumber         = "this-is-not-a-number"
	economyBulletin1   = "fwd:/economy/mybulletin1"
	financeBulletin1   = "/finance/mybulletin1"
	economyBulletin2   = "fwd:/economy/mybulletin2"
	financeBulletin2   = "/finance/mybulletin2"
	economyBulletin3   = "fwd:/economy/mybulletin3"
	financeBulletin3   = "/finance/mybulletin3"
	nonRedirectURL     = "/non-redirect-url"
	testFromURL        = "/foo"
	testToURL          = "/bar"
	expectedProto      = "https"
	expectedHost       = "api.somehost"
	expectedPathPrefix = "v1"
	idKey              = "id"
	keyValuePairs      = map[string]string{economyBulletin1: financeBulletin1, economyBulletin2: financeBulletin2, economyBulletin3: financeBulletin3}
	mockStore          = &storetest.StorerMock{
		GetKeyValuePairsFunc: func(_ context.Context, _ string, _ int64, _ uint64) (map[string]string, uint64, error) {
			return keyValuePairs, 0, nil
		},
		GetTotalKeysFunc: func(_ context.Context) (int64, error) {
			return 12, nil
		},
		GetValueFunc: func(_ context.Context, _ string) (string, error) {
			return redirectTo, nil
		},
	}
)

// encodeBase64 returns the base64 encoded string of the original URL key string
func encodeBase64(key string) string {
	encodedKey := base64.StdEncoding.EncodeToString([]byte(key))

	return encodedKey
}

func GetRedirectAPIWithMocks(backend store.Storer) *api.RedirectAPI {
	return GetRedirectAPIWithReverseLookup(backend, true)
}

func GetRedirectAPIWithReverseLookup(backend store.Storer, enableReverseLookup bool) *api.RedirectAPI {
	r := mux.NewRouter()

	baseCfg, err := config.Get()
	So(err, ShouldBeNil)

	cfg := *baseCfg
	cfg.EnablePrivateEndpoints = true
	cfg.EnableReverseLookup = enableReverseLookup
	datastoreWithConfig := store.NewDatastore(backend, &cfg)

	ctx := context.Background()
	return api.Setup(ctx, r, datastoreWithConfig, newAuthMiddlwareMock(), &cfg)
}

func TestGetRedirectEndpoint(t *testing.T) {
	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the id is valid and encoded in base64 and reverse lookup is enabled", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+existingBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, fwdPrefix+redirectFrom)
					return redirectTo, nil
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(mockStore)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the forward lookup prefixed key should be read", func() {
				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 1)
				So(getValueCalls[0].Key, ShouldEqual, fwdPrefix+redirectFrom)
			})

			Convey("Then the response status code should be 200 and the prefix should not appear in the response", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirect
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				So(response.From, ShouldEqual, validRedirect.From)
				So(response.To, ShouldEqual, validRedirect.To)
				So(response.ID, ShouldEqual, existingBase64Key)
				So(response.Links.Self.ID, ShouldEqual, existingBase64Key)
				So(response.Links.Self.Href, ShouldEqual, selfBaseURL+existingBase64Key)
			})
		})

		Convey("When the id is valid and encoded in base64 and reverse lookup is disabled", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+existingBase64Key, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, redirectFrom)
					return redirectTo, nil
				},
			}

			redirectAPI := GetRedirectAPIWithReverseLookup(mockStore, false)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the unprefixed key should be read", func() {
				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 1)
				So(getValueCalls[0].Key, ShouldEqual, redirectFrom)
			})

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirect
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				So(response.From, ShouldEqual, validRedirect.From)
				So(response.To, ShouldEqual, validRedirect.To)
				So(response.ID, ShouldEqual, existingBase64Key)
				So(response.Links.Self.ID, ShouldEqual, existingBase64Key)
				So(response.Links.Self.Href, ShouldEqual, selfBaseURL+existingBase64Key)
			})
		})
	})
}

func TestGetRedirectURLWriting(t *testing.T) {
	Convey("Given a GET /redirects/{id} request", t, func() {
		Convey("When the request headers for host, prefix and protocol are set", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectBaseURL+existingBase64Key, http.NoBody)
			request.Header.Add("X-Forwarded-Proto", expectedProto)
			request.Header.Add("X-Forwarded-Host", expectedHost)
			request.Header.Add("X-Forwarded-Path-Prefix", expectedPathPrefix)
			responseRecorder := httptest.NewRecorder()
			redirectAPI := GetRedirectAPIWithMocks(mockStore)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response body should contain the correct link", func() {
				var response models.Redirect
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response.Links.Self.ID, ShouldEqual, existingBase64Key)
				So(response.Links.Self.Href, ShouldEqual, fmt.Sprintf("%s://%s/%s/redirects/%s", expectedProto, expectedHost, expectedPathPrefix, existingBase64Key))
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
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
					return "", errors.New("key some-string not base64")
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(mockStore)
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

			// reverse lookup is enabled, so the prefixed key is the one that must miss
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, fwdPrefix+"old-path")
					return "", disRedis.ErrKeyNotFound
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(mockStore)
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
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, redirectFrom)
					return "", api.ErrInternal
				},
			}

			redirectAPI := GetRedirectAPIWithReverseLookup(mockStore, false)
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
			keyValuePairs["fwd:/economy/mybulletin4"] = "/finance/mybulletin4"
			keyValuePairs["fwd:/economy/mybulletin5"] = "/finance/mybulletin5"
			keyValuePairs["fwd:/economy/mybulletin6"] = "/finance/mybulletin6"
			keyValuePairs["fwd:/economy/mybulletin7"] = "/finance/mybulletin7"
			keyValuePairs["fwd:/economy/mybulletin8"] = "/finance/mybulletin8"
			keyValuePairs["fwd:/economy/mybulletin9"] = "/finance/mybulletin9"
			keyValuePairs["fwd:/economy/mybulletin10"] = "/finance/mybulletin10"

			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(_ context.Context, _ string, _ int64, _ uint64) (map[string]string, uint64, error) {
					return keyValuePairs, 0, nil
				},
				GetTotalKeysFunc: func(_ context.Context) (int64, error) {
					return 12, nil
				},
			}

			redirectAPI := GetRedirectAPIWithMocks(mockStore)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirects
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				respRedirectList := response.RedirectList
				respItem1 := respRedirectList[0]
				respItem1From := respItem1.From
				expectedID := encodeBase64("fwd:" + respItem1From)
				So(response.Count, ShouldEqual, 10)
				So(len(respRedirectList), ShouldEqual, 10)
				So(respItem1From, ShouldNotBeEmpty)
				So(respItem1.To, ShouldNotBeEmpty)
				So(respItem1.ID, ShouldEqual, expectedID)
				So(respItem1.Links.Self.ID, ShouldEqual, expectedID)
				SkipSo(respItem1.Links.Self.Href, ShouldEqual, getRedirectBaseURL+expectedID) // TODO change this back to 'So' when the URL rewriting functionality is fixed
				So(response.Cursor, ShouldEqual, "0")
				So(response.NextCursor, ShouldEqual, "0")
				So(response.TotalCount, ShouldEqual, 12)
			})
		})
	})
}

func TestUpsertRedirect(t *testing.T) {
	Convey("Given a valid UpsertRedirect handler", t, func() {
		mockStore := &storetest.StorerMock{
			GetValueFunc: func(_ context.Context, key string) (string, error) {
				switch key {
				case "/old-url":
					return "http://localhost:8081/new-url", nil
				case nonRedirectURL:
					return "", nil
				default:
					return "", nil
				}
			},
			SetAddFunc: func(ctx context.Context, key string, members ...interface{}) error {
				return nil
			},
			SetValueFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
				return nil
			},
		}

		apiInstance := GetRedirectAPIWithReverseLookup(mockStore, false)

		Convey("When request is valid with matching base64 ID and from path", func() {
			from := testFromURL
			to := testToURL
			id := base64.URLEncoding.EncodeToString([]byte(from))

			redirect := models.Redirect{
				From: from,
				To:   to,
			}
			body, _ := json.Marshal(redirect)
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{idKey: id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusCreated)
			So(mockStore.SetValueCalls()[0].Key, ShouldEqual, from)
		})

		Convey("When ID is not valid base64", func() {
			req := httptest.NewRequest(http.MethodPut, "/redirects/!badid", bytes.NewBuffer([]byte(`{}`)))
			req = mux.SetURLVars(req, map[string]string{idKey: "!badid"})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When base64 decodes but doesn't match body.from", func() {
			id := base64.URLEncoding.EncodeToString([]byte(testFromURL))
			redirect := models.Redirect{
				From: "/mismatch",
				To:   testToURL,
			}
			body, _ := json.Marshal(redirect)
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{idKey: id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When Redis returns an error", func() {
			mockStore.SetValueFunc = func(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
				return errors.New("redis error")
			}

			from := testFromURL
			id := base64.URLEncoding.EncodeToString([]byte(from))
			redirect := models.Redirect{
				From: from,
				To:   testToURL,
			}
			body, _ := json.Marshal(redirect)
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer(body))
			req = mux.SetURLVars(req, map[string]string{idKey: id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("When request body is invalid JSON", func() {
			id := base64.URLEncoding.EncodeToString([]byte("/foo"))
			req := httptest.NewRequest(http.MethodPut, "/redirects/"+id, bytes.NewBuffer([]byte(`{bad json`)))
			req = mux.SetURLVars(req, map[string]string{idKey: id})
			rec := httptest.NewRecorder()

			apiInstance.UpsertRedirect(rec, req)

			So(rec.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
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

			redirectAPI := GetRedirectAPIWithMocks(mockStore)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirects
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				respRedirectList := response.RedirectList
				respItem1 := respRedirectList[0]
				respItem1From := respItem1.From
				expectedID := encodeBase64("fwd:" + respItem1From)

				So(response.Count, ShouldEqual, 3)
				So(len(respRedirectList), ShouldEqual, 3)
				So(respItem1From, ShouldNotBeEmpty)
				So(respItem1.To, ShouldNotBeEmpty)
				So(respItem1.ID, ShouldEqual, expectedID)
				So(respItem1.Links.Self.ID, ShouldEqual, expectedID)
				SkipSo(respItem1.Links.Self.Href, ShouldEqual, getRedirectBaseURL+expectedID) // TODO change this back to 'So' when the URL rewriting functionality is fixed
				So(response.Cursor, ShouldEqual, "1")
				So(response.NextCursor, ShouldEqual, "0")
				So(response.TotalCount, ShouldEqual, 12)
			})
		})
	})
}

func TestGetRedirectsSuccessWithToFilter(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the to filter is valid and reverse lookup is enabled", func() {
			filterTo := financeBulletin1
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?to="+filterTo, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			mockStore := &storetest.StorerMock{
				GetSetMemberValuesFunc: func(_ context.Context, setKey, matchPattern, valuePrefix string, count int64, cursor uint64) (map[string]string, uint64, error) {
					So(setKey, ShouldEqual, "rev:"+filterTo)
					So(matchPattern, ShouldEqual, "")
					So(valuePrefix, ShouldEqual, "fwd:")
					So(count, ShouldEqual, int64(10))
					So(cursor, ShouldEqual, uint64(0))

					return map[string]string{economyBulletin1: filterTo}, 4, nil
				},
				GetSetMemberCountFunc: func(_ context.Context, setKey string) (int64, error) {
					So(setKey, ShouldEqual, "rev:"+filterTo)
					return 1, nil
				},
			}

			redirectAPI := GetRedirectAPIWithReverseLookup(mockStore, true)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 200", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				var response models.Redirects
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response.Count, ShouldEqual, 1)
				So(len(response.RedirectList), ShouldEqual, 1)
				So(response.RedirectList[0].From, ShouldEqual, economyBulletin1[4:])
				So(response.RedirectList[0].To, ShouldEqual, filterTo)
				So(response.Cursor, ShouldEqual, "0")
				So(response.NextCursor, ShouldEqual, "4")
				So(response.TotalCount, ShouldEqual, 1)
			})
		})
	})
}

func TestGetRedirectsToFilterDisabled(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the to filter is provided but reverse lookup is disabled", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?to="+financeBulletin1, http.NoBody)
			responseRecorder := httptest.NewRecorder()

			redirectAPI := GetRedirectAPIWithReverseLookup(&storetest.StorerMock{}, false)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
				So(responseRecorder.Body.String(), ShouldContainSubstring, api.ErrToNotAllowed.Error())
			})
		})
	})
}

func TestGetRedirectsInvalidToFilter(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the to filter is not a relative path", func() {
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?to=invalid-path", http.NoBody)
			responseRecorder := httptest.NewRecorder()

			redirectAPI := GetRedirectAPIWithReverseLookup(&storetest.StorerMock{}, true)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 400", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusBadRequest)
				So(responseRecorder.Body.String(), ShouldContainSubstring, api.ErrInvalidTo.Error())
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
			redirectAPI := GetRedirectAPIWithReverseLookup(mockStore, true)
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
			redirectAPI := GetRedirectAPIWithReverseLookup(mockStore, true)
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
			redirectAPI := GetRedirectAPIWithMocks(mockStore)
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
			redirectAPI := GetRedirectAPIWithMocks(mockStore)
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
				GetKeyValuePairsFunc: func(_ context.Context, _ string, _ int64, _ uint64) (map[string]string, uint64, error) {
					return nil, 0, api.ErrInternal
				},
			}
			redirectAPI := GetRedirectAPIWithMocks(mockStore)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 500", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestGetRedirectsTotalCountError(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When retrieving the total count for a filtered request fails", func() {
			filterTo := financeBulletin1
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?to="+filterTo, http.NoBody)
			responseRecorder := httptest.NewRecorder()
			mockStore := &storetest.StorerMock{
				GetSetMemberValuesFunc: func(_ context.Context, setKey, matchPattern, valuePrefix string, count int64, cursor uint64) (map[string]string, uint64, error) {
					So(setKey, ShouldEqual, "rev:"+filterTo)
					So(matchPattern, ShouldEqual, "")
					So(valuePrefix, ShouldEqual, "fwd:")
					So(count, ShouldEqual, int64(10))
					So(cursor, ShouldEqual, uint64(0))

					return map[string]string{economyBulletin1: filterTo}, 0, nil
				},
				GetSetMemberCountFunc: func(_ context.Context, setKey string) (int64, error) {
					So(setKey, ShouldEqual, "rev:"+filterTo)
					return 0, errors.New("count failed")
				},
			}

			redirectAPI := GetRedirectAPIWithReverseLookup(mockStore, true)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response status code should be 500", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestGetRedirectsURLRewriting(t *testing.T) {
	Convey("Given a GET /redirects request", t, func() {
		Convey("When the request headers for host, prefix and protocol are set", func() {
			countValue := "3"
			cursorValue := "1"
			request := httptest.NewRequest(http.MethodGet, getRedirectsBaseURL+"?count="+countValue+"&cursor="+cursorValue, http.NoBody)
			request.Header.Add("X-Forwarded-Proto", expectedProto)
			request.Header.Add("X-Forwarded-Host", expectedHost)
			request.Header.Add("X-Forwarded-Path-Prefix", expectedPathPrefix)
			responseRecorder := httptest.NewRecorder()
			redirectAPI := GetRedirectAPIWithMocks(mockStore)
			redirectAPI.Router.ServeHTTP(responseRecorder, request)

			Convey("Then the response body should contain the rewritten links", func() {
				var response models.Redirects
				err := json.Unmarshal(responseRecorder.Body.Bytes(), &response)
				So(err, ShouldBeNil)

				respRedirectList := response.RedirectList
				respItem1 := respRedirectList[0]
				respItem1From := respItem1.From
				expectedID := encodeBase64("fwd:" + respItem1From)
				So(respItem1.Links.Self.ID, ShouldEqual, expectedID)
				So(respItem1.Links.Self.Href, ShouldEqual, fmt.Sprintf("%s://%s/%s/redirects/%s", expectedProto, expectedHost, expectedPathPrefix, expectedID))
			})
		})
	})
}

func TestDeleteRedirect(t *testing.T) {
	Convey("Given a DeleteRedirect handler", t, func() {
		mockStore := &storetest.StorerMock{}

		apiInstance := GetRedirectAPIWithMocks(mockStore)

		router := mux.NewRouter()
		router.HandleFunc("/redirects/{id}", apiInstance.DeleteRedirect).Methods(http.MethodDelete)

		// Helper to encode base64
		base64ID := base64.URLEncoding.EncodeToString([]byte("/test-path"))

		Convey("When the redirect exists and is deleted successfully", func() {
			// reverse lookup is enabled, so the key must be prefixed exactly once.
			// Unmatched keys return ErrKeyNotFound, as real Redis would.
			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/test-path")
				if key == "fwd:/test-path" {
					return "/target", nil
				}
				return "", disRedis.ErrKeyNotFound
			}
			mockStore.SetRemFunc = func(_ context.Context, key string, members ...interface{}) error {
				return nil
			}
			mockStore.DeleteValueFunc = func(_ context.Context, key string) error {
				return nil
			}

			req := httptest.NewRequest(http.MethodDelete, "/redirects/"+base64ID, http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusNoContent)
		})

		Convey("When the redirect does not exist", func() {
			mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
				return "", disRedis.ErrKeyNotFound
			}

			req := httptest.NewRequest(http.MethodDelete, "/redirects/"+base64ID, http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			So(rr.Code, ShouldEqual, http.StatusNotFound)
			So(rr.Body.String(), ShouldContainSubstring, "not found")
		})

		Convey("When an internal error occurs during existence check", func() {
			mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
				return "", errors.New("connection failed")
			}

			req := httptest.NewRequest(http.MethodDelete, "/redirects/"+base64ID, http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusInternalServerError)
			So(rr.Body.String(), ShouldContainSubstring, "internal error")
		})

		Convey("When the base64 id is invalid", func() {
			req := httptest.NewRequest(http.MethodDelete, "/redirects/invalid_base64", http.NoBody)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			So(rr.Code, ShouldEqual, http.StatusBadRequest)
			So(rr.Body.String(), ShouldContainSubstring, "the base64 id provided is invalid")
		})
	})
}
