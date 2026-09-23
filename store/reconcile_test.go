package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestReverseLookupReconcilerReconcile(t *testing.T) {
	Convey("Given a reverse lookup reconciler with redirects to reconcile", t, func() {
		Convey("When reconcile scans multiple batches", func() {
			ctx := context.Background()
			mockStore := &storetest.StorerMock{}

			mockStore.GetKeysFunc = func(_ context.Context, matchPattern string, count int64, cursor uint64) ([]string, uint64, error) {
				So(matchPattern, ShouldEqual, "/*")
				So(count, ShouldEqual, int64(100))

				switch cursor {
				case 0:
					return []string{"/old-1"}, 12, nil
				case 12:
					return []string{"/old-2"}, 0, nil
				default:
					return nil, 0, errors.New("unexpected cursor")
				}
			}

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				switch key {
				case "/old-1":
					return "/new-1", nil
				case "/old-2":
					return "/new-2", nil
				default:
					return "", errors.New("unexpected key")
				}
			}

			mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
				So(queue, ShouldNotBeNil)
				return nil, nil
			}

			reconciler := store.NewReverseLookupReconciler(store.Datastore{Backend: mockStore})
			changes, err := reconciler.Reconcile(ctx)

			Convey("Then it converts each redirect correctly", func() {
				So(err, ShouldBeNil)
				So(changes, ShouldEqual, 2)

				getKeyCalls := mockStore.GetKeysCalls()
				So(getKeyCalls, ShouldHaveLength, 2)
				So(getKeyCalls[0].MatchPattern, ShouldEqual, "/*")
				So(getKeyCalls[0].Count, ShouldEqual, int64(100))
				So(getKeyCalls[0].Cursor, ShouldEqual, uint64(0))
				So(getKeyCalls[1].MatchPattern, ShouldEqual, "/*")
				So(getKeyCalls[1].Count, ShouldEqual, int64(100))
				So(getKeyCalls[1].Cursor, ShouldEqual, uint64(12))

				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 2)
				So(getValueCalls[0].Key, ShouldEqual, "/old-1")
				So(getValueCalls[1].Key, ShouldEqual, "/old-2")

				transactionCalls := mockStore.TransactionCalls()
				So(transactionCalls, ShouldHaveLength, 2)
			})
		})

		Convey("When reconcile fails during processing", func() {
			testCases := []struct {
				name          string
				configureMock func(*storetest.StorerMock, error)
				wantChanges   int
				wantErr       string
				errMessage    string
			}{
				{
					name: "scan failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return nil, 27, expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=27: get redirects with match failed: scan failed",
					errMessage:  "scan failed",
				},
				{
					name: "get value failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "", expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to reverse lookup failed: get original redirect value failed: get failed",
					errMessage:  "get failed",
				},
				{
					name: "transaction failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
							So(queue, ShouldNotBeNil)
							return nil, expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to reverse lookup failed: transaction failed: transaction failed",
					errMessage:  "transaction failed",
				},
			}

			for _, tc := range testCases {
				tc := tc
				Convey(tc.name, func() {
					mockStore := &storetest.StorerMock{}
					expectedErr := errors.New(tc.errMessage)
					tc.configureMock(mockStore, expectedErr)

					reconciler := store.NewReverseLookupReconciler(store.Datastore{Backend: mockStore})
					changes, err := reconciler.Reconcile(context.Background())

					So(changes, ShouldEqual, tc.wantChanges)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, tc.wantErr)
				})
			}
		})
	})
}

func TestForwardLookupOnlyReconcilerReconcile(t *testing.T) {
	Convey("Given a forward lookup only reconciler with prefixed redirects to reconcile", t, func() {
		Convey("When reconcile scans multiple batches", func() {
			ctx := context.Background()
			mockStore := &storetest.StorerMock{}

			mockStore.GetKeysFunc = func(_ context.Context, matchPattern string, count int64, cursor uint64) ([]string, uint64, error) {
				So(matchPattern, ShouldEqual, "fwd:*")
				So(count, ShouldEqual, int64(100))

				switch cursor {
				case 0:
					return []string{"fwd:/old-1"}, 9, nil
				case 9:
					return []string{"fwd:/old-2"}, 0, nil
				default:
					return nil, 0, errors.New("unexpected cursor")
				}
			}

			mockStore.GetValueFunc = func(_ context.Context, key string) (string, error) {
				switch key {
				case "fwd:/old-1":
					return "/new-1", nil
				case "fwd:/old-2":
					return "/new-2", nil
				default:
					return "", errors.New("unexpected key")
				}
			}

			mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
				So(queue, ShouldNotBeNil)
				return nil, nil
			}

			reconciler := store.NewForwardLookupOnlyReconciler(store.Datastore{Backend: mockStore})
			changes, err := reconciler.Reconcile(ctx)

			Convey("Then it converts each prefixed redirect correctly", func() {
				So(err, ShouldBeNil)
				So(changes, ShouldEqual, 2)

				getKeyCalls := mockStore.GetKeysCalls()
				So(getKeyCalls, ShouldHaveLength, 2)
				So(getKeyCalls[0].MatchPattern, ShouldEqual, "fwd:*")
				So(getKeyCalls[0].Count, ShouldEqual, int64(100))
				So(getKeyCalls[0].Cursor, ShouldEqual, uint64(0))
				So(getKeyCalls[1].MatchPattern, ShouldEqual, "fwd:*")
				So(getKeyCalls[1].Count, ShouldEqual, int64(100))
				So(getKeyCalls[1].Cursor, ShouldEqual, uint64(9))

				getValueCalls := mockStore.GetValueCalls()
				So(getValueCalls, ShouldHaveLength, 2)
				So(getValueCalls[0].Key, ShouldEqual, "fwd:/old-1")
				So(getValueCalls[1].Key, ShouldEqual, "fwd:/old-2")

				transactionCalls := mockStore.TransactionCalls()
				So(transactionCalls, ShouldHaveLength, 2)
			})
		})

		Convey("When reconcile fails during processing", func() {
			testCases := []struct {
				name          string
				configureMock func(*storetest.StorerMock, error)
				wantChanges   int
				wantErr       string
				errMessage    string
			}{
				{
					name: "scan failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return nil, 33, expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=33: get redirects with match failed: scan failed",
					errMessage:  "scan failed",
				},
				{
					name: "get value failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"fwd:/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "", expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to forward lookup only failed: get original redirect value failed: get failed",
					errMessage:  "get failed",
				},
				{
					name: "transaction failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"fwd:/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.TransactionFunc = func(_ context.Context, queue func(redis.Pipeliner)) ([]redis.Cmder, error) {
							So(queue, ShouldNotBeNil)
							return nil, expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to forward lookup only failed: transaction failed: transaction failed",
					errMessage:  "transaction failed",
				},
			}

			for _, tc := range testCases {
				tc := tc
				Convey(tc.name, func() {
					mockStore := &storetest.StorerMock{}
					expectedErr := errors.New(tc.errMessage)
					tc.configureMock(mockStore, expectedErr)

					reconciler := store.NewForwardLookupOnlyReconciler(store.Datastore{Backend: mockStore})
					changes, err := reconciler.Reconcile(context.Background())

					So(changes, ShouldEqual, tc.wantChanges)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, tc.wantErr)
				})
			}
		})
	})
}
