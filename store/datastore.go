package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ONSdigital/log.go/v2/log"

	disRedis "github.com/ONSdigital/dis-redis"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/redis/go-redis/v9"
)

const fwdPrefix = "fwd:"
const revPrefix = "rev:"

//go:generate moq -out datastoretest/redis.go -pkg storetest . Redis
//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

// NewDatastore creates a new instance of Datastore
// with the provided backend and configuration.
func NewDatastore(backend Storer, cfg *config.Config) *Datastore {
	return &Datastore{
		Backend: backend,
		cfg:     cfg,
	}
}

// Datastore represents a generic data store that abstracts the
// underlying storage backend.
type Datastore struct {
	Backend Storer
	cfg     *config.Config
}

type dataRedis interface {
	Checker(ctx context.Context, state *healthcheck.CheckState) error
	GetValue(ctx context.Context, key string) (string, error)
	GetKeyValuePairs(ctx context.Context, matchPattern string, count int64, cursor uint64) (keyValuePairs map[string]string, newCursor uint64, err error)
	GetTotalKeys(ctx context.Context) (totalKeys int64, err error)
	GetSetMemberValues(ctx context.Context, setKey, matchPattern, valuePrefix string, count int64, cursor uint64) (map[string]string, uint64, error)
	GetSetMemberCount(ctx context.Context, setKey string) (count int64, err error)
	SetValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetAdd(ctx context.Context, key string, members ...interface{}) error
	SetRem(ctx context.Context, key string, members ...interface{}) error
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
// datastore based on the provided redirectID. When reverse lookup is
// enabled the redirect is stored under the forward lookup prefix, so
// the key is prefixed before reading. The error is returned unwrapped
// so that callers can compare it against disRedis.ErrKeyNotFound.
func (ds *Datastore) GetRedirect(ctx context.Context, redirectID string) (string, error) {
	if ds.cfg.EnableReverseLookup {
		fwdRedirectKey := fmt.Sprintf("%s%s", fwdPrefix, redirectID)
		return ds.Backend.GetValue(ctx, fwdRedirectKey)
	}

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

	var matchPattern = ""
	enableReverseLookup := ds.cfg.EnableReverseLookup

	if enableReverseLookup {
		matchPattern = "fwd*"
	}
	var allKeyValuePairs map[string]string
	// Use a count of 1000 to bring back up to the maximum number of redirects before we count the number we need - otherwise this won't
	// bring back the correct number if using a matchPattern e.g. fwd*
	allKeyValuePairs, newCursor, err = ds.Backend.GetKeyValuePairs(ctx, matchPattern, 1000, cursor)
	if err != nil {
		log.Error(ctx, "failed to get key value pairs from redis", err)
		return nil, 0, err
	}

	// Loop through the returned redirects and count the required number into the keyValuePairs map
	var i int64 = 0
	keyValuePairs = make(map[string]string)
	for key, value := range allKeyValuePairs {
		if i == count {
			break
		}
		keyValuePairs[key] = value
		i++
	}

	if enableReverseLookup {
		for key := range keyValuePairs {
			// check the index of the first instance of "fwd:" in key; will give -1 if substr is not present
			i := strings.Index(key, "fwd:")
			if i == 0 {
				cutKey := key[4:]
				keyValuePairs[cutKey] = keyValuePairs[key]
				delete(keyValuePairs, key)
			}
		}
	}
	return keyValuePairs, newCursor, err
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
// provided redirectID. This is the raw key accessor and deliberately
// applies no prefix, regardless of the reverse lookup flag: its callers
// either pass an already prefixed key or intend to read the unprefixed
// one. Use GetRedirect for prefix aware redirect lookups.
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
//
// Deprecated: Use UpsertRedirect instead.
func (ds *Datastore) UpsertValue(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return ds.Backend.SetValue(ctx, key, value, expiration)
}

// UpsertRedirect inserts or updates a redirect in the datastore
// with the specified 'from' and 'to' values.
func (ds *Datastore) UpsertRedirect(ctx context.Context, from, to string) error {
	if ds.cfg.EnableReverseLookup {
		fwdRedirectKey := fmt.Sprintf("%s%s", fwdPrefix, from)
		revRedirectKey := fmt.Sprintf("%s%s", revPrefix, to)

		// We need the old value in case this is an update.
		oldValue, err := ds.Backend.GetValue(ctx, fwdRedirectKey)
		if err != nil && err != disRedis.ErrKeyNotFound {
			return fmt.Errorf("failed to get old redirect value: %w", err)
		}

		oldRevRedirectKey := fmt.Sprintf("%s%s", revPrefix, oldValue)

		if err := ds.Backend.SetAdd(ctx, revRedirectKey, from); err != nil {
			return fmt.Errorf("failed to add reverse lookup: %w", err)
		}
		if err := ds.Backend.SetValue(ctx, fwdRedirectKey, to, 0); err != nil {
			return fmt.Errorf("failed to set redirect value: %w", err)
		}
		if oldValue != "" {
			if err := ds.Backend.SetRem(ctx, oldRevRedirectKey, from); err != nil {
				return fmt.Errorf("failed to remove old reverse lookup: %w", err)
			}
		}

		return nil
	}

	return ds.Backend.SetValue(ctx, from, to, 0)
}

// DeleteValue removes a value from the datastore based on the
// provided redirectID.
//
// Deprecated: Use DeleteRedirect instead.
func (ds *Datastore) DeleteValue(ctx context.Context, redirectID string) error {
	return ds.Backend.DeleteValue(ctx, redirectID)
}

// DeleteRedirect removes a redirect and its reverse lookup entry from the
// datastore based on the provided redirectID.
func (ds *Datastore) DeleteRedirect(ctx context.Context, redirectID string) error {
	if ds.cfg.EnableReverseLookup {
		fwdRedirectKey := fwdPrefix + redirectID
		// GetValue, not GetRedirect: the key is already prefixed here, and
		// GetRedirect would prefix it a second time.
		to, err := ds.Backend.GetValue(ctx, fwdRedirectKey)
		if err != nil {
			return fmt.Errorf("failed to get redirect value: %w", err)
		}

		if err := ds.Backend.SetRem(ctx, revPrefix+to, redirectID); err != nil {
			return fmt.Errorf("failed to remove reverse lookup: %w", err)
		}
		if err := ds.Backend.DeleteValue(ctx, fwdRedirectKey); err != nil && err != disRedis.ErrKeyNotFound {
			return fmt.Errorf("failed to delete redirect value: %w", err)
		}

		return nil
	}

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

	if err := ds.Backend.SetValue(ctx, fwdRedirectKey, value, 0); err != nil {
		return fmt.Errorf("set forward redirect failed: %w", err)
	}
	if err := ds.Backend.SetAdd(ctx, revRedirectKey, key); err != nil {
		return fmt.Errorf("add reverse lookup failed: %w", err)
	}
	if err := ds.Backend.DeleteValue(ctx, key); err != nil && err != disRedis.ErrKeyNotFound {
		return fmt.Errorf("delete original redirect failed: %w", err)
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

	if err := ds.Backend.SetValue(ctx, fwdRedirectKey, value, 0); err != nil {
		return fmt.Errorf("set forward redirect failed: %w", err)
	}
	if err := ds.Backend.SetRem(ctx, revRedirectKey, fwdRedirectKey); err != nil {
		return fmt.Errorf("remove reverse lookup failed: %w", err)
	}
	if err := ds.Backend.DeleteValue(ctx, key); err != nil && err != disRedis.ErrKeyNotFound {
		return fmt.Errorf("delete original redirect failed: %w", err)
	}
	return nil
}
