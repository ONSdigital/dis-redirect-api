// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package storetest

import (
	"context"
	"github.com/ONSdigital/dis-redirect-api/store"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"sync"
)

// Ensure, that RedisMock does implement store.Redis.
// If this is not the case, regenerate this file with moq.
var _ store.Redis = &RedisMock{}

// RedisMock is a mock implementation of store.Redis.
//
//	func TestSomethingThatUsesRedis(t *testing.T) {
//
//		// make and configure a mocked store.Redis
//		mockedRedis := &RedisMock{
//			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
//				panic("mock out the Checker method")
//			},
//			GetValueFunc: func(ctx context.Context, key string) (string, error) {
//				panic("mock out the GetValue method")
//			},
//		}
//
//		// use mockedRedis in code that requires store.Redis
//		// and then make assertions.
//
//	}
type RedisMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// GetValueFunc mocks the GetValue method.
	GetValueFunc func(ctx context.Context, key string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// GetValue holds details about calls to the GetValue method.
		GetValue []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Key is the key argument value.
			Key string
		}
	}
	lockChecker  sync.RWMutex
	lockGetValue sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *RedisMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("RedisMock.CheckerFunc: method is nil but Redis.Checker was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(contextMoqParam, checkState)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//
//	len(mockedRedis.CheckerCalls())
func (mock *RedisMock) CheckerCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// GetValue calls GetValueFunc.
func (mock *RedisMock) GetValue(ctx context.Context, key string) (string, error) {
	if mock.GetValueFunc == nil {
		panic("RedisMock.GetValueFunc: method is nil but Redis.GetValue was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Key string
	}{
		Ctx: ctx,
		Key: key,
	}
	mock.lockGetValue.Lock()
	mock.calls.GetValue = append(mock.calls.GetValue, callInfo)
	mock.lockGetValue.Unlock()
	return mock.GetValueFunc(ctx, key)
}

// GetValueCalls gets all the calls that were made to GetValue.
// Check the length with:
//
//	len(mockedRedis.GetValueCalls())
func (mock *RedisMock) GetValueCalls() []struct {
	Ctx context.Context
	Key string
} {
	var calls []struct {
		Ctx context.Context
		Key string
	}
	mock.lockGetValue.RLock()
	calls = mock.calls.GetValue
	mock.lockGetValue.RUnlock()
	return calls
}
