package sdk

import (
	"context"

	"github.com/ONSdigital/dis-redirect-api/models"
	apiError "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

// Clienter is an interface defining the methods available for
// interacting with the redirect API.
//
//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter
type Clienter interface {
	URL() string
	Health() *healthcheck.Client
	Checker(ctx context.Context, check *health.CheckState) error
	GetRedirect(ctx context.Context, options Options, key string) (*models.Redirect, apiError.Error)
	GetRedirects(ctx context.Context, options Options) (*models.Redirects, apiError.Error)
	PutRedirect(ctx context.Context, options Options, id string, payload models.Redirect) apiError.Error
	DeleteRedirect(ctx context.Context, options Options, id string) apiError.Error
}
