package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/api.go -pkg apimock . Redis

// Redis defines the required methods for Redis
type Redis interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}
