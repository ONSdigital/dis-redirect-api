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
	RedirectEndpoint = "%s/v1/redirects/%s"
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
