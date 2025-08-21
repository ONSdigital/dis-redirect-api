package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	apiError "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	service = "dis-redirect-api"
)

type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given redirect api url
func NewClient(redirectAPIURL string) *Client {
	return &Client{
		hcCli: healthcheck.NewClient(service, redirectAPIURL),
	}
}

// NewWithHealthClient creates a new instance of redirect API Client,
// reusing the URL and Clienter from the provided healthcheck client
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		hcCli: healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (cli *Client) URL() string {
	return cli.hcCli.URL
}

// Health returns the underlying Healthcheck Client for this redirect API client
func (cli *Client) Health() *healthcheck.Client {
	return cli.hcCli
}

// Checker calls redirect api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

type ResponseInfo struct {
	Body    []byte
	Headers http.Header
	Status  int
}

// callRedirectAPI calls the Redirect API endpoint given by path for the provided REST method, request headers, and body payload.
// It returns the response body and any error that occurred.
func (cli *Client) callRedirectAPI(ctx context.Context, path, method string, headers http.Header, queryParams url.Values, payload []byte) (*ResponseInfo, apiError.Error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to parse path: \"%v\" error is: %v", path, err),
		}
	}

	path = URL.String()

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload))
	} else {
		req, err = http.NewRequest(method, path, http.NoBody)
	}

	// check req, above, didn't error
	if err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to create request for call to redirect api, error is: %v", err),
		}
	}

	// set any headers against request
	setHeaders(req, headers)

	if payload != nil {
		req.Header.Add("Content-type", "application/json")
	}

	// add the count and cursor values, if present, to the path URL
	q := req.URL.Query()
	q.Add("count", queryParams.Get("count"))
	q.Add("cursor", queryParams.Get("cursor"))
	req.URL.RawQuery = q.Encode()

	resp, err := cli.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, apiError.StatusError{
			Err:  fmt.Errorf("failed to call redirect api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}
	defer func() {
		err = closeResponseBody(resp)
	}()

	respInfo := &ResponseInfo{
		Headers: resp.Header.Clone(),
		Status:  resp.StatusCode,
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 400 {
		return respInfo, apiError.StatusError{
			Err:  fmt.Errorf("failed as unexpected code from redirect api: %v", resp.StatusCode),
			Code: resp.StatusCode,
		}
	}

	if resp.Body == nil {
		return respInfo, nil
	}

	respInfo.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return respInfo, apiError.StatusError{
			Err:  fmt.Errorf("failed to read response body from call to redirect api, error is: %v", err),
			Code: resp.StatusCode,
		}
	}
	return respInfo, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(resp *http.Response) apiError.Error {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			return apiError.StatusError{
				Err:  fmt.Errorf("error closing http response body from call to redirect api, error is: %v", err),
				Code: http.StatusInternalServerError,
			}
		}
	}

	return nil
}
