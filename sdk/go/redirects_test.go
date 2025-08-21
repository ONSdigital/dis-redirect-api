package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var getRedirectsResponse = models.Redirects{}

func TestGetRedirect(t *testing.T) {
	t.Parallel()

	headers := http.Header{}

	Convey("Given a request to get redirect", t, func() {
		body, err := json.Marshal(getRedirectResponse)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		Convey("When GetRedirect is called", func() {
			resp, err := redirectAPIClient.GetRedirect(ctx, Options{Headers: headers}, existingBase64Key)

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, getRedirectResponse)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, fmt.Sprintf("/v1/redirects/%s", existingBase64Key))
					})
				})
			})
		})
	})
}

func TestGetRedirects(t *testing.T) {
	t.Parallel()

	headers := http.Header{}
	queryValues := url.Values{}

	Convey("Given a request to get the default number of redirects", t, func() {
		body, err := json.Marshal(getRedirectsResponse)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		Convey("When GetRedirects is called using the default values of count and cursor", func() {
			queryValues.Set("count", "")
			queryValues.Set("cursor", "")
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers, Query: queryValues})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, getRedirectsResponse)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects")
						So(doCalls[0].Req.URL.Query().Get("count"), ShouldEqual, "")
						So(doCalls[0].Req.URL.Query().Get("cursor"), ShouldEqual, "")
					})
				})
			})
		})

		Convey("When GetRedirects is called using specific values of count and cursor", func() {
			queryValues.Set("count", "2")
			queryValues.Set("cursor", "1")
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers, Query: queryValues})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, getRedirectsResponse)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects")
						So(doCalls[0].Req.URL.Query().Get("count"), ShouldEqual, "2")
						So(doCalls[0].Req.URL.Query().Get("cursor"), ShouldEqual, "1")
					})
				})
			})
		})

		Convey("When GetRedirects is called using a specific value of count only", func() {
			queryValues.Set("count", "3")
			queryValues.Set("cursor", "")
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers, Query: queryValues})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, getRedirectsResponse)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects")
						So(doCalls[0].Req.URL.Query().Get("count"), ShouldEqual, "3")
						So(doCalls[0].Req.URL.Query().Get("cursor"), ShouldEqual, "")
					})
				})
			})
		})

		Convey("When GetRedirects is called using a specific value of cursor only", func() {
			queryValues.Set("count", "")
			queryValues.Set("cursor", "2")
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers, Query: queryValues})

			Convey("Then the expected response body is returned", func() {
				So(*resp, ShouldResemble, getRedirectsResponse)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects")
						So(doCalls[0].Req.URL.Query().Get("count"), ShouldEqual, "")
						So(doCalls[0].Req.URL.Query().Get("cursor"), ShouldEqual, "2")
					})
				})
			})
		})
	})
}

func TestPutRedirect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	headers := http.Header{
		Authorization: {"Bearer authorised-user"},
	}

	Convey("Given a successful 201 Created response from dis-redirect-api", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       nil,
				Header:     nil,
			},
			nil,
		)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		redirect := models.Redirect{
			From: "/old-url",
			To:   "/new-url",
		}

		Convey("When PutRedirect is called", func() {
			err := redirectAPIClient.PutRedirect(ctx, Options{Headers: headers}, "L29sZC11cmw=", redirect) // base64(/old-url)

			Convey("Then it succeeds with no errors returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And client.Do should be called once with the expected URL", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects/L29sZC11cmw=") // URL encoded base64
				So(doCalls[0].Req.Method, ShouldEqual, http.MethodPut)
			})
		})
	})

	Convey("Given a 500 Internal Server Error from dis-redirect-api", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			nil,
		)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)
		redirect := models.Redirect{From: "/broken", To: "/fail"}

		Convey("When PutRedirect is called", func() {
			err := redirectAPIClient.PutRedirect(ctx, Options{Headers: headers}, "L2Jyb2tlbg==", redirect)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusInternalServerError)
			})

			Convey("And client.Do is called with the correct path", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects/L2Jyb2tlbg==")
			})
		})
	})

	Convey("Given an unexpected client error occurs", t, func() {
		clientErr := errors.New("network error")
		httpClient := newMockHTTPClient(nil, clientErr)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)
		redirect := models.Redirect{From: "/error", To: "/nowhere"}

		Convey("When PutRedirect is called", func() {
			err := redirectAPIClient.PutRedirect(ctx, Options{Headers: headers}, "L2Vycm9y", redirect)

			Convey("Then a wrapped error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "network error")
			})

			Convey("And client.Do is called once", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
			})
		})
	})
}

func TestDeleteRedirect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	headers := http.Header{
		Authorization: {"Bearer authorised-user"},
	}

	Convey("Given a successful 204 No Content response from dis-redirect-api", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusNoContent,
				Body:       nil,
				Header:     nil,
			},
			nil,
		)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		Convey("When DeleteRedirect is called", func() {
			err := redirectAPIClient.DeleteRedirect(ctx, Options{Headers: headers}, "L29sZC11cmw=") // base64(/old-url)

			Convey("Then it succeeds with no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And client.Do should be called with correct path and method", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, "/v1/redirects/L29sZC11cmw=")
				So(doCalls[0].Req.Method, ShouldEqual, http.MethodDelete)
			})
		})
	})

	Convey("Given a 404 Not Found response from dis-redirect-api", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusNotFound,
			},
			nil,
		)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		Convey("When DeleteRedirect is called", func() {
			err := redirectAPIClient.DeleteRedirect(ctx, Options{Headers: headers}, "L25vdC1mb3VuZA==") // base64(/not-found)

			Convey("Then a not found error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusNotFound)
			})
		})
	})

	Convey("Given a 500 Internal Server Error response", t, func() {
		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusInternalServerError,
			},
			nil,
		)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		Convey("When DeleteRedirect is called", func() {
			err := redirectAPIClient.DeleteRedirect(ctx, Options{Headers: headers}, "L2ZhaWx1cmU=")

			Convey("Then an internal server error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Status(), ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given a network error occurs", t, func() {
		netErr := errors.New("connection reset by peer")
		httpClient := newMockHTTPClient(nil, netErr)

		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		Convey("When DeleteRedirect is called", func() {
			err := redirectAPIClient.DeleteRedirect(ctx, Options{Headers: headers}, "L25ldC1mYWls")

			Convey("Then a wrapped error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "connection reset by peer")
			})
		})
	})
}
