package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dis-redirect-api/api"
	apiError "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
)

const (
	RedirectEndpoint = "%s/redirects/%s"
)

// GetRedirect gets the /redirects/{id} endpoint
func (cli *Client) GetRedirect(ctx context.Context, options Options, key string) (*api.RedirectResponse, apiError.Error) {
	path := fmt.Sprintf(RedirectEndpoint, cli.hcCli.URL, key)

	respInfo, apiErr := cli.callRedirectAPI(ctx, path, http.MethodGet, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var response api.RedirectResponse

	if err := json.Unmarshal(respInfo.Body, &response); err != nil {
		return nil, apiError.StatusError{
			Err: fmt.Errorf("failed to unmarshal getRedirect response - error is: %v", err),
		}
	}

	return &response, nil
}
