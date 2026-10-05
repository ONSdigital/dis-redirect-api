package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
	disRedis "github.com/ONSdigital/dis-redis"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDatastoreUpsertRedirect(t *testing.T) {
	Convey("Given a datastore upserting redirects", t, func() {
		ctx := context.Background()

		Convey("When reverse lookup is disabled", func() {
			mockStore := &storetest.StorerMock{
				SetValueFunc: func(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
					return nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: false})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the redirect should be upserted correctly", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldBeEmpty)
				So(mockStore.SetAddCalls(), ShouldBeEmpty)

				setValueCalls := mockStore.SetValueCalls()
				So(setValueCalls, ShouldHaveLength, 1)
				So(setValueCalls[0].Key, ShouldEqual, "/old")
				So(setValueCalls[0].Value, ShouldEqual, "/new")
				So(setValueCalls[0].Expiration, ShouldEqual, time.Duration(0))
			})
		})

		Convey("When reverse lookup is enabled for a new redirect", func() {
			mockStore := &storetest.StorerMock{}

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/old")
				return "", disRedis.ErrKeyNotFound
			}
			mockStore.SetAddFunc = func(_ context.Context, _ string, _ ...interface{}) error {
				return nil
			}
			mockStore.SetValueFunc = func(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
				return nil
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the redirect should be upserted correctly with reverse lookup", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldHaveLength, 1)
				So(mockStore.SetAddCalls(), ShouldHaveLength, 1)
				So(mockStore.SetAddCalls()[0].Key, ShouldEqual, "rev:/new")
				So(mockStore.SetAddCalls()[0].Members, ShouldResemble, []interface{}{"/old"})

				So(mockStore.SetValueCalls(), ShouldHaveLength, 1)
				So(mockStore.SetValueCalls()[0].Key, ShouldEqual, "fwd:/old")
				So(mockStore.SetValueCalls()[0].Value, ShouldEqual, "/new")
				So(mockStore.SetValueCalls()[0].Expiration, ShouldEqual, time.Duration(0))
			})
		})

		Convey("When reverse lookup is enabled for an existing redirect", func() {
			mockStore := &storetest.StorerMock{}

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/old")
				return "/previous", nil
			}

			mockStore.SetAddFunc = func(_ context.Context, _ string, _ ...interface{}) error {
				return nil
			}

			mockStore.SetValueFunc = func(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
				return nil
			}

			mockStore.SetRemFunc = func(_ context.Context, _ string, _ ...interface{}) error {
				return nil
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the redirect should be upserted correctly with reverse lookup", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldHaveLength, 1)
				So(mockStore.SetAddCalls(), ShouldHaveLength, 1)
				So(mockStore.SetAddCalls()[0].Key, ShouldEqual, "rev:/new")
				So(mockStore.SetAddCalls()[0].Members, ShouldResemble, []interface{}{"/old"})

				So(mockStore.SetValueCalls(), ShouldHaveLength, 1)
				So(mockStore.SetValueCalls()[0].Key, ShouldEqual, "fwd:/old")
				So(mockStore.SetValueCalls()[0].Value, ShouldEqual, "/new")
				So(mockStore.SetValueCalls()[0].Expiration, ShouldEqual, time.Duration(0))

				So(mockStore.SetRemCalls(), ShouldHaveLength, 1)
				So(mockStore.SetRemCalls()[0].Key, ShouldEqual, "rev:/previous")
				So(mockStore.SetRemCalls()[0].Members, ShouldResemble, []interface{}{"/old"})
			})
		})

		Convey("When looking up the existing reverse mapping fails", func() {
			expectedErr := errors.New("get failed")
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
					return "", expectedErr
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the error should be returned and no transaction or set value should occur", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to get old redirect value: get failed")
				So(mockStore.SetAddCalls(), ShouldBeEmpty)
				So(mockStore.SetValueCalls(), ShouldBeEmpty)
			})
		})

		Convey("When adding the reverse lookup fails", func() {
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
					return "", disRedis.ErrKeyNotFound
				},
				SetAddFunc: func(_ context.Context, _ string, _ ...interface{}) error {
					return errors.New("failed to add reverse lookup")
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the error should be returned and no set value should occur", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to add reverse lookup: failed to add reverse lookup")
				So(mockStore.SetValueCalls(), ShouldBeEmpty)
			})
		})
	})
}

func TestDatastoreGetRedirect(t *testing.T) {
	Convey("Given a datastore retrieving a redirect", t, func() {
		ctx := context.Background()

		Convey("When reverse lookup is disabled", func() {
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, "/old")
					return "/new", nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: false})
			value, err := datastore.GetRedirect(ctx, "/old")

			Convey("Then the unprefixed key should be read", func() {
				So(err, ShouldBeNil)
				So(value, ShouldEqual, "/new")

				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 1)
				So(getValueCalls[0].Key, ShouldEqual, "/old")
			})
		})

		Convey("When reverse lookup is enabled", func() {
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, "fwd:/old")
					return "/new", nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			value, err := datastore.GetRedirect(ctx, "/old")

			Convey("Then the forward lookup prefixed key should be read", func() {
				So(err, ShouldBeNil)
				So(value, ShouldEqual, "/new")

				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 1)
				So(getValueCalls[0].Key, ShouldEqual, "fwd:/old")
			})
		})

		Convey("When reverse lookup is enabled and the key does not exist", func() {
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, "fwd:/old")
					return "", disRedis.ErrKeyNotFound
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			value, err := datastore.GetRedirect(ctx, "/old")

			Convey("Then the error should be returned unwrapped so callers can compare it", func() {
				So(err, ShouldEqual, disRedis.ErrKeyNotFound)
				So(value, ShouldBeEmpty)
			})
		})
	})
}

func TestDatastoreGetValue(t *testing.T) {
	Convey("Given a datastore getting a raw value with reverse lookup enabled", t, func() {
		ctx := context.Background()

		mockStore := &storetest.StorerMock{
			GetValueFunc: func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "/old")
				return "/new", nil
			},
		}

		datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
		value, err := datastore.GetValue(ctx, "/old")

		Convey("Then the key should not be prefixed, unlike GetRedirect", func() {
			So(err, ShouldBeNil)
			So(value, ShouldEqual, "/new")

			getValueCalls := mockStore.GetValueCalls()
			So(getValueCalls, ShouldHaveLength, 1)
			So(getValueCalls[0].Key, ShouldEqual, "/old")
		})
	})
}

func TestDatastoreDeleteRedirect(t *testing.T) {
	Convey("Given a datastore deleting redirects", t, func() {
		ctx := context.Background()

		Convey("When reverse lookup is disabled", func() {
			mockStore := &storetest.StorerMock{
				DeleteValueFunc: func(_ context.Context, key string) error {
					So(key, ShouldEqual, "/old")
					return nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: false})
			err := datastore.DeleteRedirect(ctx, "/old")

			Convey("Then the unprefixed key should be deleted directly", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldBeEmpty)
				So(mockStore.SetRemCalls(), ShouldBeEmpty)

				deleteValueCalls := mockStore.DeleteValueCalls()
				So(deleteValueCalls, ShouldHaveLength, 1)
				So(deleteValueCalls[0].Key, ShouldEqual, "/old")
			})
		})

		Convey("When reverse lookup is enabled", func() {
			mockStore := &storetest.StorerMock{}

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/old")
				return "/new", nil
			}

			mockStore.SetRemFunc = func(_ context.Context, key string, members ...interface{}) error {
				So(key, ShouldEqual, "rev:/new")
				So(members, ShouldResemble, []interface{}{"/old"})
				return nil
			}

			mockStore.DeleteValueFunc = func(_ context.Context, key string) error {
				So(key, ShouldEqual, "fwd:/old")
				return nil
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.DeleteRedirect(ctx, "/old")

			Convey("Then the key should be prefixed exactly once and the reverse entry removed", func() {
				So(err, ShouldBeNil)

				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 1)
				So(getValueCalls[0].Key, ShouldEqual, "fwd:/old")

				setRemCalls := mockStore.SetRemCalls()
				So(setRemCalls, ShouldHaveLength, 1)
				So(setRemCalls[0].Key, ShouldEqual, "rev:/new")
				So(setRemCalls[0].Members, ShouldResemble, []interface{}{"/old"})

				deleteValueCalls := mockStore.DeleteValueCalls()
				So(deleteValueCalls, ShouldHaveLength, 1)
				So(deleteValueCalls[0].Key, ShouldEqual, "fwd:/old")
			})
		})

		Convey("When reverse lookup is enabled and the redirect does not exist", func() {
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, "fwd:/old")
					return "", disRedis.ErrKeyNotFound
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.DeleteRedirect(ctx, "/old")

			Convey("Then a not found error should be returned and nothing deleted", func() {
				So(err, ShouldNotBeNil)
				So(errors.Is(err, disRedis.ErrKeyNotFound), ShouldBeTrue)
				So(mockStore.SetRemCalls(), ShouldBeEmpty)
				So(mockStore.DeleteValueCalls(), ShouldBeEmpty)
			})
		})
	})
}

func TestDatastoreGetRedirects(t *testing.T) {
	Convey("Given a datastore retrieving redirects", t, func() {
		ctx := context.Background()
		const count int64 = 25
		const cursor uint64 = 4

		Convey("When reverse lookup is disabled", func() {
			var receivedPattern string
			var receivedCount int64
			var receivedCursor uint64

			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(
					_ context.Context,
					matchPattern string,
					requestedCount int64,
					requestedCursor uint64,
				) (map[string]string, uint64, error) {
					receivedPattern = matchPattern
					receivedCount = requestedCount
					receivedCursor = requestedCursor
					return map[string]string{"/old": "/new"}, 9, nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: false})
			redirects, newCursor, err := datastore.GetRedirects(ctx, "", count, cursor)

			Convey("Then it should return redirects and pass through pagination arguments", func() {
				So(err, ShouldBeNil)
				So(redirects, ShouldResemble, map[string]string{"/old": "/new"})
				So(newCursor, ShouldEqual, 9)
				So(receivedPattern, ShouldBeEmpty)
				So(receivedCount, ShouldEqual, count)
				So(receivedCursor, ShouldEqual, cursor)
				So(mockStore.GetKeyValuePairsCalls(), ShouldHaveLength, 1)
				So(mockStore.GetSetMemberValuesCalls(), ShouldBeEmpty)
			})
		})

		Convey("When reverse lookup is enabled", func() {
			var receivedPattern string

			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(
					_ context.Context,
					matchPattern string,
					_ int64,
					_ uint64,
				) (map[string]string, uint64, error) {
					receivedPattern = matchPattern
					return map[string]string{
						"fwd:/old":   "/new",
						"fwd:/other": "/destination",
					}, 12, nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			redirects, newCursor, err := datastore.GetRedirects(ctx, "", count, cursor)

			Convey("Then it should filter forward keys and remove their prefixes", func() {
				So(err, ShouldBeNil)
				So(redirects, ShouldResemble, map[string]string{
					"/old":   "/new",
					"/other": "/destination",
				})
				So(newCursor, ShouldEqual, 12)
				So(receivedPattern, ShouldEqual, "fwd*")
				So(mockStore.GetKeyValuePairsCalls(), ShouldHaveLength, 1)
				So(mockStore.GetSetMemberValuesCalls(), ShouldBeEmpty)
			})
		})

		Convey("When a destination is provided", func() {
			var receivedSetKey string
			var receivedPattern string
			var receivedPrefix string
			var receivedCount int64
			var receivedCursor uint64

			mockStore := &storetest.StorerMock{
				GetSetMemberValuesFunc: func(
					_ context.Context,
					setKey string,
					matchPattern string,
					valuePrefix string,
					requestedCount int64,
					requestedCursor uint64,
				) (map[string]string, uint64, error) {
					receivedSetKey = setKey
					receivedPattern = matchPattern
					receivedPrefix = valuePrefix
					receivedCount = requestedCount
					receivedCursor = requestedCursor
					return map[string]string{"/old": "/destination"}, 6, nil
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			redirects, newCursor, err := datastore.GetRedirects(ctx, "/destination", count, cursor)

			Convey("Then it should retrieve members from the destination reverse index", func() {
				So(err, ShouldBeNil)
				So(redirects, ShouldResemble, map[string]string{"/old": "/destination"})
				So(newCursor, ShouldEqual, 6)
				So(receivedSetKey, ShouldEqual, "rev:/destination")
				So(receivedPattern, ShouldBeEmpty)
				So(receivedPrefix, ShouldEqual, "fwd:")
				So(receivedCount, ShouldEqual, count)
				So(receivedCursor, ShouldEqual, cursor)
				So(mockStore.GetSetMemberValuesCalls(), ShouldHaveLength, 1)
				So(mockStore.GetKeyValuePairsCalls(), ShouldBeEmpty)
			})
		})

		Convey("When retrieving all redirects fails", func() {
			expectedErr := errors.New("lookup failed")
			mockStore := &storetest.StorerMock{
				GetKeyValuePairsFunc: func(
					_ context.Context,
					_ string,
					_ int64,
					_ uint64,
				) (map[string]string, uint64, error) {
					return map[string]string{"/old": "/new"}, 3, expectedErr
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: false})
			redirects, newCursor, err := datastore.GetRedirects(ctx, "", count, cursor)

			Convey("Then it should return an error to say that the lookup failed", func() {
				So(err, ShouldEqual, expectedErr)
				So(redirects, ShouldResemble, map[string]string(nil))
				So(newCursor, ShouldEqual, 0)
			})
		})
	})
}
