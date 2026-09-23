package store

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out datastoretest/redis.go -pkg storetest . Redis
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

// Datastore represents a generic data store that abstracts the
// underlying storage backend.
type Datastore struct {
	Backend Storer
}

type dataRedis interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetValue(ctx context.Context, key string) (string, error)
	GetKeyValuePairs(ctx context.Context, matchPattern string, count int64, cursor uint64) (keyValuePairs map[string]string, newCursor uint64, err error)
	GetTotalKeys(ctx context.Context) (totalKeys int64, err error)
	SetValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeleteValue(ctx context.Context, key string) error
}

// Redis represents all the required methods from Redis
type Redis interface {
	dataRedis
	Checker(context.Context, *healthcheck.CheckState) error
}

// Storer represents basic data access via Get, Remove and
// Upsert methods, abstracting it from Redis
type Storer interface {
	dataRedis
}

// GetRedirect retrieves a redirect value from the
// datastore based on the provided redirectID.
func (ds *Datastore) GetRedirect(ctx context.Context, redirectID string) (string, error) {
	return ds.Backend.GetValue(ctx, redirectID)
}

// GetRedirects retrieves a set of key-value pairs from the datastore
// based on the specified count and cursor for pagination.
func (ds *Datastore) GetRedirects(ctx context.Context, count int64, cursor uint64) (keyValuePairs map[string]string, newCursor uint64, err error) {
	return ds.Backend.GetKeyValuePairs(ctx, "", count, cursor)
}

// GetTotalCount retrieves the total number of keys in the datastore.
func (ds *Datastore) GetTotalCount(ctx context.Context) (totalCount int, err error) {
	var totalKeys int64
	totalKeys, err = ds.Backend.GetTotalKeys(ctx)
	if err != nil {
		return -1, err
	}
	totalCount = int(totalKeys)
	return totalCount, err
}

// GetValue retrieves a value from the datastore based on the
// provided redirectID.
func (ds *Datastore) GetValue(ctx context.Context, redirectID string) (string, error) {
	return ds.Backend.GetValue(ctx, redirectID)
}

// UpsertValue inserts or updates a value in the datastore with the
// specified key, value, and expiration time.
func (ds *Datastore) UpsertValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return ds.Backend.SetValue(ctx, key, value, expiration)
}

// DeleteValue removes a value from the datastore based on the
// provided redirectID.
func (ds *Datastore) DeleteValue(ctx context.Context, redirectID string) error {
	return ds.Backend.DeleteValue(ctx, redirectID)
}
