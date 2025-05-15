package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dis-redirect-api/api"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	c "github.com/smartystreets/goconvey/convey"
)

const testHost = "http://localhost:23900"

var (
	initialTestState = healthcheck.CreateCheckState(service)

	helloWorldResponse = api.HelloResponse{
		Message: "hello there",
	}
)

func TestHealthCheckerClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	timePriorHealthCheck := time.Now().UTC()
	path := "/health"

	c.Convey("Given clienter.Do returns an error", t, func() {
		clientError := errors.New("unexpected error")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		redirectAPIClient := newRedirectAPIClient(t, httpClient)
		check := initialTestState

		c.Convey("When redirect API client Checker is called", func() {
			err := redirectAPIClient.Checker(ctx, &check)
			c.So(err, c.ShouldBeNil)

			c.Convey("Then the expected check is returned", func() {
				c.So(check.Name(), c.ShouldEqual, service)
				c.So(check.Status(), c.ShouldEqual, health.StatusCritical)
				c.So(check.StatusCode(), c.ShouldEqual, 0)
				c.So(check.Message(), c.ShouldEqual, clientError.Error())
				c.So(*check.LastChecked(), c.ShouldHappenAfter, timePriorHealthCheck)
				c.So(check.LastSuccess(), c.ShouldBeNil)
				c.So(*check.LastFailure(), c.ShouldHappenAfter, timePriorHealthCheck)
			})

			c.Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				c.So(doCalls, c.ShouldHaveLength, 1)
				c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, path)
			})
		})
	})

	c.Convey("Given a 500 response for health check", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		redirectAPIClient := newRedirectAPIClient(t, httpClient)
		check := initialTestState

		c.Convey("When redirect API client Checker is called", func() {
			err := redirectAPIClient.Checker(ctx, &check)
			c.So(err, c.ShouldBeNil)

			c.Convey("Then the expected check is returned", func() {
				c.So(check.Name(), c.ShouldEqual, service)
				c.So(check.Status(), c.ShouldEqual, health.StatusCritical)
				c.So(check.StatusCode(), c.ShouldEqual, 500)
				c.So(check.Message(), c.ShouldEqual, service+healthcheck.StatusMessage[health.StatusCritical])
				c.So(*check.LastChecked(), c.ShouldHappenAfter, timePriorHealthCheck)
				c.So(check.LastSuccess(), c.ShouldBeNil)
				c.So(*check.LastFailure(), c.ShouldHappenAfter, timePriorHealthCheck)
			})

			c.Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				c.So(doCalls, c.ShouldHaveLength, 1)
				c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, path)
			})
		})
	})
}

func TestHelloWorld(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	headers := http.Header{
		Authorization: {"Bearer authorised-user"},
	}

	c.Convey("Given request is authorised to say hello", t, func() {
		body, err := json.Marshal(helloWorldResponse)
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

		c.Convey("When GetHelloWorld is called", func() {
			resp, err := redirectAPIClient.GetHelloWorld(ctx, Options{Headers: headers})

			c.Convey("Then the expected response body is returned", func() {
				c.So(*resp, c.ShouldResemble, helloWorldResponse)

				c.Convey("And no error is returned", func() {
					c.So(err, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/hello")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldResemble, []string{"Bearer authorised-user"})
					})
				})
			})
		})
	})

	c.Convey("Given a 401 response from redirect api", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusUnauthorized}, nil)
		redirectAPIClient := newRedirectAPIClient(t, httpClient)

		c.Convey("When GetHelloWorld is called", func() {
			resp, err := redirectAPIClient.GetHelloWorld(ctx, Options{})

			c.Convey("Then an error should be returned ", func() {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Status(), c.ShouldEqual, http.StatusUnauthorized)
				c.So(err.Error(), c.ShouldEqual, "failed as unexpected code from redirect api: 401")

				c.Convey("And the expected response body should be nil", func() {
					c.So(resp, c.ShouldBeNil)

					c.Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						c.So(doCalls, c.ShouldHaveLength, 1)
						c.So(doCalls[0].Req.Method, c.ShouldEqual, "GET")
						c.So(doCalls[0].Req.URL.Path, c.ShouldEqual, "/hello")
						c.So(doCalls[0].Req.Header["Authorization"], c.ShouldBeEmpty)
					})
				})
			})
		})
	})
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newRedirectAPIClient(_ *testing.T, httpClient *dphttp.ClienterMock) *Client {
	healthClient := healthcheck.NewClientWithClienter(service, testHost, httpClient)
	return NewWithHealthClient(healthClient)
}
