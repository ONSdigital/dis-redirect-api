package store

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out datastoretest/redis.go -pkg storetest . Redis
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

type Datastore struct {
	Backend Storer
}

type dataRedis interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetValue(ctx context.Context, key string) (string, error)
}

// Redis represents all the required methods from Redis
type Redis interface {
	dataRedis
	Checker(context.Context, *healthcheck.CheckState) error
}

// Storer represents basic data access via Get, Remove and Upsert methods, abstracting it from Redis
type Storer interface {
	dataRedis
}

func (ds *Datastore) GetRedirect(ctx context.Context, redirectID string) (string, error) {
	return ds.Backend.GetValue(ctx, redirectID)
}
