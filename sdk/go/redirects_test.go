package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

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
