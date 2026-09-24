package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/redis/go-redis/v9"
)

const fwdPrefix = "fwd:"
const revPrefix = "rev:"

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
	GetSetMemberValues(ctx context.Context, setKey, matchPattern, valuePrefix string, count int64, cursor uint64) (map[string]string, uint64, error)
	GetSetMemberCount(ctx context.Context, setKey string) (count int64, err error)
	SetValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetKeys(ctx context.Context, matchPattern string, count int64, cursor uint64) (keys []string, newCursor uint64, err error)
	Transaction(ctx context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error)
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

// GetRedirects retrieves a list of redirects from the datastore.
// If the 'to' parameter is provided, it retrieves redirects
// pointing to the specified destination.
// Otherwise, it retrieves all redirects with pagination support.
func (ds *Datastore) GetRedirects(ctx context.Context, to string, count int64, cursor uint64) (keyValuePairs map[string]string, newCursor uint64, err error) {
	if to != "" {
		return ds.Backend.GetSetMemberValues(ctx, revPrefix+to, "", fwdPrefix, count, cursor)
	}

	return ds.Backend.GetKeyValuePairs(ctx, "", count, cursor)
}

// GetTotalCount retrieves the total count of redirects from the datastore.
// If the 'to' parameter is provided, it retrieves the count of
// redirects pointing to the specified destination.
// Otherwise, it retrieves the total count of all redirects.
func (ds *Datastore) GetTotalCount(ctx context.Context, to string) (totalCount int, err error) {
	if to != "" {
		var setMemberCount int64
		setMemberCount, err = ds.Backend.GetSetMemberCount(ctx, revPrefix+to)
		if err != nil {
			return -1, err
		}
		totalCount = int(setMemberCount)
		return totalCount, err
	}
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

// GetKeys retrieves a set of keys from the datastore
// based on the specified match pattern, count, and cursor for pagination.
func (ds *Datastore) GetKeys(ctx context.Context, matchPattern string, count int64, cursor uint64) (keys []string, newCursor uint64, err error) {
	return ds.Backend.GetKeys(ctx, matchPattern, count, cursor)
}

// GetSetMemberValues retrieves the values of members in a set
// from the datastore based on the provided set key
// and optional match pattern and value prefix.
func (ds *Datastore) GetSetMemberValues(ctx context.Context, setKey, matchPattern, valuePrefix string, count int64, cursor uint64) (keyValuePairs map[string]string, newCursor uint64, err error) {
	return ds.Backend.GetSetMemberValues(ctx, setKey, matchPattern, valuePrefix, count, cursor)
}

// UpsertValue inserts or updates a value in the
// datastore with the specified key, value, and expiration time.
func (ds *Datastore) UpsertValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return ds.Backend.SetValue(ctx, key, value, expiration)
}

// DeleteValue removes a value from the datastore based on the
// provided redirectID.
func (ds *Datastore) DeleteValue(ctx context.Context, redirectID string) error {
	return ds.Backend.DeleteValue(ctx, redirectID)
}

// ConvertRedirectFormatToReverseLookup converts a
// forward lookup redirect format to a dual index
// format allowing reverse lookup in the datastore.
func (ds *Datastore) ConvertRedirectFormatToReverseLookup(ctx context.Context, key string) error {
	value, err := ds.GetValue(ctx, key)
	if err != nil {
		return fmt.Errorf("get original redirect value failed: %w", err)
	}

	fwdRedirectKey := fmt.Sprintf("%s%s", fwdPrefix, key)
	revRedirectKey := fmt.Sprintf("%s%s", revPrefix, value)

	_, err = ds.Backend.Transaction(ctx, func(pipe redis.Pipeliner) {
		pipe.Set(ctx, fwdRedirectKey, value, 0)
		pipe.SAdd(ctx, revRedirectKey, key)
		pipe.Del(ctx, key)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}

// ConvertRedirectFormatToForwardLookupOnly converts a
// dual index redirect format to a forward lookup only format in the datastore.
func (ds *Datastore) ConvertRedirectFormatToForwardLookupOnly(ctx context.Context, key string) error {
	value, err := ds.GetValue(ctx, key)
	if err != nil {
		return fmt.Errorf("get original redirect value failed: %w", err)
	}

	fwdRedirectKey := strings.TrimPrefix(key, fwdPrefix)
	revRedirectKey := fmt.Sprintf("%s%s", revPrefix, value)

	_, err = ds.Backend.Transaction(ctx, func(pipe redis.Pipeliner) {
		pipe.Set(ctx, fwdRedirectKey, value, 0)
		pipe.SRem(ctx, revRedirectKey, fwdRedirectKey)
		pipe.Del(ctx, key)
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}
