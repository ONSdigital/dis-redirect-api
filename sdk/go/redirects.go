package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dis-redirect-api/models"
	apiError "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
)

const (
	RedirectEndpoint  = "%s/v1/redirects/%s"
	RedirectsEndpoint = "%s/v1/redirects"
)

// GetRedirect gets the /redirects/{id} endpoint
func (cli *Client) GetRedirect(ctx context.Context, options Options, key string) (*models.Redirect, apiError.Error) {
	path := fmt.Sprintf(RedirectEndpoint, cli.hcCli.URL, key)

	respInfo, apiErr := cli.callRedirectAPI(ctx, path, http.MethodGet, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var response models.Redirect
	if err := json.Unmarshal(respInfo.Body, &response); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal redirect response - error is: %v", err),
		}
	}

	return &response, nil
}

// GetRedirects gets the /redirects endpoint
func (cli *Client) GetRedirects(ctx context.Context, options Options, count, cursor string) (*models.Redirects, apiError.Error) {
	path := fmt.Sprintf(RedirectsEndpoint, cli.hcCli.URL)

	if count != "" {
		path = path + "?count=" + count
		if cursor != "" {
			path = path + "&cursor=" + cursor
		}
	} else if cursor != "" {
		path = path + "?cursor=" + cursor
	}

	respInfo, apiErr := cli.callRedirectAPI(ctx, path, http.MethodGet, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var response models.Redirects
	if err := json.Unmarshal(respInfo.Body, &response); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal redirect response - error is: %v", err),
		}
	}

	return &response, nil
}
