package sdk

import (
	"context"
	"errors"
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
	getRedirectResponse = api.RedirectResponse{
		Value: "new-path",
	}
	initialTestState  = healthcheck.CreateCheckState(service)
	existingBase64Key = "Y29va2llLXNhdWNl"
	ctx               = context.Background()
)

func TestHealthCheckerClient(t *testing.T) {
	t.Parallel()

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
