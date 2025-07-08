package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/models"
	c "github.com/smartystreets/goconvey/convey"
)

var getRedirectsResponse = models.Redirects{}

func TestGetRedirect(t *testing.T) {
	t.Parallel()

	headers := http.Header{}

	c.Convey("Given a request to get redirect", t, func() {
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

		c.Convey("When GetRedirect is called", func() {
			resp, err := redirectAPIClient.GetRedirect(ctx, Options{Headers: headers}, existingBase64Key)

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, getRedirectResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, fmt.Sprintf("/v1/redirects/%s", existingBase64Key))
					})
				})
			})
		})
	})
}

func TestGetRedirects(t *testing.T) {
	t.Parallel()

	headers := http.Header{}

	c.Convey("Given a request to get the default number of redirects", t, func() {
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

		c.Convey("When GetRedirects is called using the default values of count and cursor", func() {
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers}, "", "")

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, getRedirectsResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/v1/redirects")
						c.So(doCalls[0].Req.URL.Query().Get("count"), c.ShouldEqual, "")
						c.So(doCalls[0].Req.URL.Query().Get("cursor"), c.ShouldEqual, "")
					})
				})
			})
		})

		c.Convey("When GetRedirects is called using specific values of count and cursor", func() {
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers}, "2", "1")

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, getRedirectsResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/v1/redirects")
						c.So(doCalls[0].Req.URL.Query().Get("count"), c.ShouldEqual, "2")
						c.So(doCalls[0].Req.URL.Query().Get("cursor"), c.ShouldEqual, "1")
					})
				})
			})
		})

		c.Convey("When GetRedirects is called using a specific value of count only", func() {
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers}, "3", "")

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, getRedirectsResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/v1/redirects")
						c.So(doCalls[0].Req.URL.Query().Get("count"), c.ShouldEqual, "3")
						c.So(doCalls[0].Req.URL.Query().Get("cursor"), c.ShouldEqual, "")
					})
				})
			})
		})

		c.Convey("When GetRedirects is called using a specific value of cursor only", func() {
			resp, err := redirectAPIClient.GetRedirects(ctx, Options{Headers: headers}, "", "2")

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, getRedirectsResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/v1/redirects")
						c.So(doCalls[0].Req.URL.Query().Get("count"), c.ShouldEqual, "")
						c.So(doCalls[0].Req.URL.Query().Get("cursor"), c.ShouldEqual, "2")
					})
				})
			})
		})
	})
}
