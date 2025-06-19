package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestGetRedirect(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		Authorization: {"Bearer authorised-user"},
	}

	c.Convey("Given request is authorised to get redirect", t, func() {
		body, err := json.Marshal(getRedirectResponse)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
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
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, fmt.Sprintf("/redirects/%s", existingBase64Key))
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})
}
