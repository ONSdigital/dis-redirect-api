package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-redirect-api/store"
	storetest "github.com/ONSdigital/dis-redirect-api/store/datastoretest"
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

			mockStore.SetValueFunc = func(_ context.Context, key string, value interface{}, expiration time.Duration) error {
				return nil
			}

			mockStore.SetAddFunc = func(_ context.Context, key string, members ...interface{}) error {
				return nil
			}

			mockStore.DeleteValueFunc = func(_ context.Context, key string) error {
				return nil
			}

			reconciler := store.NewReverseLookupReconciler(&store.Datastore{Backend: mockStore})
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

				setValueCalls := mockStore.SetValueCalls()
				So(setValueCalls, ShouldHaveLength, 2)
				So(setValueCalls[0].Key, ShouldEqual, "fwd:/old-1")
				So(setValueCalls[1].Key, ShouldEqual, "fwd:/old-2")

				setAddCalls := mockStore.SetAddCalls()
				So(setAddCalls, ShouldHaveLength, 2)
				So(setAddCalls[0].Key, ShouldEqual, "rev:/new-1")
				So(setAddCalls[0].Members, ShouldResemble, []interface{}{"/old-1"})
				So(setAddCalls[1].Key, ShouldEqual, "rev:/new-2")
				So(setAddCalls[1].Members, ShouldResemble, []interface{}{"/old-2"})
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
					name: "set value failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.SetValueFunc = func(_ context.Context, key string, value interface{}, expiration time.Duration) error {
							return expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to reverse lookup failed: set forward redirect failed: set failed",
					errMessage:  "set failed",
				},
				{
					name: "set add failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.SetValueFunc = func(_ context.Context, key string, value interface{}, expiration time.Duration) error {
							return nil
						}
						mockStore.SetAddFunc = func(_ context.Context, key string, members ...interface{}) error {
							return expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to reverse lookup failed: add reverse lookup failed: set add failed",
					errMessage:  "set add failed",
				},
				{
					name: "delete value failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.SetValueFunc = func(_ context.Context, key string, value interface{}, expiration time.Duration) error {
							return nil
						}
						mockStore.SetAddFunc = func(_ context.Context, key string, members ...interface{}) error {
							return nil
						}
						mockStore.DeleteValueFunc = func(_ context.Context, key string) error {
							return expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to reverse lookup failed: delete original redirect failed: delete failed",
					errMessage:  "delete failed",
				},
			}

			for _, tc := range testCases {
				tc := tc
				Convey(tc.name, func() {
					mockStore := &storetest.StorerMock{}
					expectedErr := errors.New(tc.errMessage)
					tc.configureMock(mockStore, expectedErr)

					reconciler := store.NewReverseLookupReconciler(&store.Datastore{Backend: mockStore})
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

			mockStore.SetRemFunc = func(_ context.Context, key string, members ...interface{}) error {
				return nil
			}

			mockStore.SetValueFunc = func(_ context.Context, key string, value interface{}, expiration time.Duration) error {
				return nil
			}

			mockStore.DeleteValueFunc = func(_ context.Context, key string) error {
				return nil
			}

			reconciler := store.NewForwardLookupOnlyReconciler(&store.Datastore{Backend: mockStore})
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

				setRemCalls := mockStore.SetRemCalls()
				So(setRemCalls, ShouldHaveLength, 2)
				So(setRemCalls[0].Key, ShouldEqual, "rev:/new-1")
				So(setRemCalls[0].Members, ShouldResemble, []interface{}{"/old-1"})

				setValueCalls := mockStore.SetValueCalls()
				So(setValueCalls, ShouldHaveLength, 2)
				So(setValueCalls[0].Key, ShouldEqual, "/old-1")
				So(setValueCalls[0].Value, ShouldEqual, "/new-1")

				deleteValueCalls := mockStore.DeleteValueCalls()
				So(deleteValueCalls, ShouldHaveLength, 2)
				So(deleteValueCalls[0].Key, ShouldEqual, "fwd:/old-1")
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
					name: "set value failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"fwd:/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.SetValueFunc = func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
							return expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to forward lookup only failed: set forward redirect failed: set value failed",
					errMessage:  "set value failed",
				},
				{
					name: "set rem failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"fwd:/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.SetValueFunc = func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
							return nil
						}
						mockStore.SetRemFunc = func(ctx context.Context, key string, members ...interface{}) error {
							return expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to forward lookup only failed: remove reverse lookup failed: set rem failed",
					errMessage:  "set rem failed",
				},
				{
					name: "delete value failure",
					configureMock: func(mockStore *storetest.StorerMock, expectedErr error) {
						mockStore.GetKeysFunc = func(_ context.Context, _ string, _ int64, _ uint64) ([]string, uint64, error) {
							return []string{"fwd:/old"}, 0, nil
						}
						mockStore.GetValueFunc = func(_ context.Context, _ string) (string, error) {
							return "/new", nil
						}
						mockStore.SetValueFunc = func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
							return nil
						}
						mockStore.SetRemFunc = func(ctx context.Context, key string, members ...interface{}) error {
							return nil
						}
						mockStore.DeleteValueFunc = func(ctx context.Context, key string) error {
							return expectedErr
						}
					},
					wantChanges: 0,
					wantErr:     "reconcile batch scan failed: cursor=0: convert redirect format to forward lookup only failed: delete original redirect failed: delete value failed",
					errMessage:  "delete value failed",
				},
			}

			for _, tc := range testCases {
				tc := tc
				Convey(tc.name, func() {
					mockStore := &storetest.StorerMock{}
					expectedErr := errors.New(tc.errMessage)
					tc.configureMock(mockStore, expectedErr)

					reconciler := store.NewForwardLookupOnlyReconciler(&store.Datastore{Backend: mockStore})
					changes, err := reconciler.Reconcile(context.Background())

					So(changes, ShouldEqual, tc.wantChanges)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, tc.wantErr)
				})
			}
		})
	})
}
