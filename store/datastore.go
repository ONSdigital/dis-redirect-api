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
	GetKeyValuePairs(ctx context.Context, matchPattern string, count int64, cursor uint64) (keyValuePairs map[string]string, newCursor uint64, err error)
	GetTotalKeys(ctx context.Context) (totalKeys int64, err error)
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

func (ds *Datastore) GetRedirects(ctx context.Context, count int64, cursor uint64) (map[string]string, uint64, error) {
	return ds.Backend.GetKeyValuePairs(ctx, "", count, cursor)
}

func (ds *Datastore) GetTotalCount(ctx context.Context) (totalCount int, err error) {
	var totalKeys int64
	totalKeys, err = ds.Backend.GetTotalKeys(ctx)
	if err != nil {
		return -1, err
	}
	totalCount = int(totalKeys)
	return totalCount, err
}
