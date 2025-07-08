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

var (
	//redirect1Self = models.RedirectSelf{
	//	Href: "https://api.beta.ons.gov.uk/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgx",
	//	Id:   "L2Vjb25vbXkvb2xkLXBhdGgx",
	//}
	//redirect1Links = models.RedirectLinks{
	//	Self: redirect1Self,
	//}
	//redirect1 = models.Redirect{
	//	From:  "/economy/old-path1",
	//	To:    "/economy/new-path1",
	//	Id:    "L2Vjb25vbXkvb2xkLXBhdGgx",
	//	Links: redirect1Links,
	//}
	//redirect2Self = models.RedirectSelf{
	//	Href: "https://api.beta.ons.gov.uk/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgy",
	//	Id:   "L2Vjb25vbXkvb2xkLXBhdGgy",
	//}
	//redirect2Links = models.RedirectLinks{
	//	Self: redirect2Self,
	//}
	//redirect2 = models.Redirect{
	//	From:  "/economy/old-path2",
	//	To:    "/economy/new-path2",
	//	Id:    "L2Vjb25vbXkvb2xkLXBhdGgy",
	//	Links: redirect2Links,
	//}
	//redirect3Self = models.RedirectSelf{
	//	Href: "https://api.beta.ons.gov.uk/v1/redirects/L2Vjb25vbXkvb2xkLXBhdGgz",
	//	Id:   "L2Vjb25vbXkvb2xkLXBhdGgz",
	//}
	//redirect3Links = models.RedirectLinks{
	//	Self: redirect3Self,
	//}
	//redirect3 = models.Redirect{
	//	From:  "/economy/old-path3",
	//	To:    "/economy/new-path3",
	//	Id:    "L2Vjb25vbXkvb2xkLXBhdGgz",
	//	Links: redirect3Links,
	//}
	redirectList         = make([]models.Redirect, 0, 2)
	getRedirectsResponse = models.Redirects{
		Count:        10,
		RedirectList: redirectList,
		Cursor:       "0",
		TotalCount:   3,
	}
)

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

		c.Convey("When GetRedirect is called", func() {
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
					})
				})
			})
		})
	})
}
