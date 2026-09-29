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
	"github.com/redis/go-redis/v9"
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
				So(mockStore.TransactionCalls(), ShouldBeEmpty)

				setValueCalls := mockStore.SetValueCalls()
				So(setValueCalls, ShouldHaveLength, 1)
				So(setValueCalls[0].Key, ShouldEqual, "/old")
				So(setValueCalls[0].Value, ShouldEqual, "/new")
				So(setValueCalls[0].Expiration, ShouldEqual, time.Duration(0))
			})
		})

		Convey("When reverse lookup is enabled for a new redirect", func() {
			mockStore := &storetest.StorerMock{}
			var queuedCommands []redis.Cmder

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/old")
				return "", disRedis.ErrKeyNotFound
			}
			mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
				client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
				defer client.Close()

				pipe := client.Pipeline()
				queue(pipe)
				queuedCommands = append([]redis.Cmder(nil), pipe.Cmds()...)

				return nil, nil
			}
			mockStore.SetValueFunc = func(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
				return nil
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the redirect should be upserted correctly with reverse lookup", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldHaveLength, 1)
				So(mockStore.TransactionCalls(), ShouldHaveLength, 1)
				So(queuedCommands, ShouldHaveLength, 2)
				So(queuedCommands[0].Name(), ShouldEqual, "set")
				So(queuedCommands[0].Args()[1], ShouldEqual, "fwd:/old")
				So(queuedCommands[0].Args()[2], ShouldEqual, "/new")
				So(queuedCommands[1].Name(), ShouldEqual, "sadd")
				So(queuedCommands[1].Args()[1], ShouldEqual, "rev:/new")
				So(queuedCommands[1].Args()[2], ShouldEqual, "/old")
			})
		})

		Convey("When reverse lookup is enabled for an existing redirect", func() {
			mockStore := &storetest.StorerMock{}
			var queuedCommands []redis.Cmder

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/old")
				return "/previous", nil
			}
			mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
				client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
				defer client.Close()

				pipe := client.Pipeline()
				queue(pipe)
				queuedCommands = append([]redis.Cmder(nil), pipe.Cmds()...)

				return nil, nil
			}
			mockStore.SetValueFunc = func(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
				return nil
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the redirect should be upserted correctly with reverse lookup", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldHaveLength, 1)
				So(mockStore.TransactionCalls(), ShouldHaveLength, 1)
				So(queuedCommands, ShouldHaveLength, 3)
				So(queuedCommands[0].Name(), ShouldEqual, "set")
				So(queuedCommands[0].Args()[1], ShouldEqual, "fwd:/old")
				So(queuedCommands[0].Args()[2], ShouldEqual, "/new")
				So(queuedCommands[1].Name(), ShouldEqual, "sadd")
				So(queuedCommands[1].Args()[1], ShouldEqual, "rev:/new")
				So(queuedCommands[1].Args()[2], ShouldEqual, "/old")
				So(queuedCommands[2].Name(), ShouldEqual, "srem")
				So(queuedCommands[2].Args()[1], ShouldEqual, "rev:/previous")
				So(queuedCommands[2].Args()[2], ShouldEqual, "/old")
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
				So(mockStore.TransactionCalls(), ShouldBeEmpty)
				So(mockStore.SetValueCalls(), ShouldBeEmpty)
			})
		})

		Convey("When the reverse lookup transaction fails", func() {
			expectedErr := errors.New("transaction failed")
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, _ string) (string, error) {
					return "", disRedis.ErrKeyNotFound
				},
				TransactionFunc: func(_ context.Context, _ func(redis.Pipeliner)) ([]redis.Cmder, error) {
					return nil, expectedErr
				},
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.UpsertRedirect(ctx, "/old", "/new")

			Convey("Then the error should be returned and no set value should occur", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "transaction failed: transaction failed")
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

			redirect, err := datastore.GetRedirect(ctx, "/old")

			Convey("Then the unprefixed redirect should be retrieved", func() {
				So(err, ShouldBeNil)
				So(redirect, ShouldEqual, "/new")
			})
		})

		Convey("When reverse lookup is enabled", func() {
			mockStore := &storetest.StorerMock{
				GetValueFunc: func(_ context.Context, key string) (string, error) {
					So(key, ShouldEqual, "/old")
					return "/new", nil
				},
			}
			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})

			redirect, err := datastore.GetRedirect(ctx, "/old")

			Convey("Then the redirect key should not be modified", func() {
				So(err, ShouldBeNil)
				So(redirect, ShouldEqual, "/new")
			})
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

			Convey("Then the unprefixed redirect should be deleted", func() {
				So(err, ShouldBeNil)
				So(mockStore.GetValueCalls(), ShouldBeEmpty)
				So(mockStore.TransactionCalls(), ShouldBeEmpty)
				So(mockStore.DeleteValueCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("When reverse lookup is enabled", func() {
			mockStore := &storetest.StorerMock{}
			var queuedCommands []redis.Cmder

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				So(key, ShouldEqual, "fwd:/old")
				return "/new", nil
			}
			mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
				client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
				defer client.Close()

				pipe := client.Pipeline()
				queue(pipe)
				queuedCommands = append([]redis.Cmder(nil), pipe.Cmds()...)

				return nil, nil
			}

			datastore := store.NewDatastore(mockStore, &config.Config{EnableReverseLookup: true})
			err := datastore.DeleteRedirect(ctx, "/old")

			Convey("Then the forward redirect and reverse set member should be removed atomically", func() {
				So(err, ShouldBeNil)
				So(mockStore.DeleteValueCalls(), ShouldBeEmpty)
				So(mockStore.TransactionCalls(), ShouldHaveLength, 1)
				So(queuedCommands, ShouldHaveLength, 2)
				So(queuedCommands[0].Name(), ShouldEqual, "del")
				So(queuedCommands[0].Args()[1], ShouldEqual, "fwd:/old")
				So(queuedCommands[1].Name(), ShouldEqual, "srem")
				So(queuedCommands[1].Args()[1], ShouldEqual, "rev:/new")
				So(queuedCommands[1].Args()[2], ShouldEqual, "/old")
			})
		})
	})
}
