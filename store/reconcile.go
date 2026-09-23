package store

import (
	"context"
	"fmt"
)

// Reconciler defines the interface for a component that can
// reconcile stored records.
type Reconciler interface {
	Reconcile(ctx context.Context) (int, error)
}

// ReverseLookupReconciler is a reconciler that ensures reverse lookup
// keys are correctly maintained for stored redirects.
type ReverseLookupReconciler struct {
	ds Datastore
}

// NewReverseLookupReconciler creates a new instance of ReverseLookupReconciler.
func NewReverseLookupReconciler(ds Datastore) *ReverseLookupReconciler {
	return &ReverseLookupReconciler{
		ds: ds,
	}
}

// Reconcile will convert all stored redirects to use fwd prefix and
// add a reverse lookup via the rev prefix
func (r *ReverseLookupReconciler) Reconcile(ctx context.Context) (int, error) {
	return reconcileBatch(ctx, r.reconcileBatchForReverseLookup)
}

func (r *ReverseLookupReconciler) reconcileBatchForReverseLookup(ctx context.Context, cursor uint64) (recordsChanged int, newCursor uint64, err error) {
	redirectKeys, newCursor, err := r.ds.GetKeys(ctx, "/*", 100, cursor)
	if err != nil {
		return recordsChanged, newCursor, fmt.Errorf("get redirects with match failed: %w", err)
	}

	for _, key := range redirectKeys {
		err := r.ds.ConvertRedirectFormatToReverseLookup(ctx, key)
		if err != nil {
			return recordsChanged, newCursor, fmt.Errorf("convert redirect format to reverse lookup failed: %w", err)
		}
		recordsChanged++
	}

	return recordsChanged, newCursor, nil
}

// ForwardLookupOnlyReconciler is a reconciler that ensures only forward lookup
// keys are maintained for stored redirects.
type ForwardLookupOnlyReconciler struct {
	ds Datastore
}

// NewForwardLookupOnlyReconciler creates a new instance of
// ForwardLookupOnlyReconciler.
func NewForwardLookupOnlyReconciler(ds Datastore) *ForwardLookupOnlyReconciler {
	return &ForwardLookupOnlyReconciler{
		ds: ds,
	}
}

// Reconcile will convert all stored redirects to use only forward
// lookup keys and remove any reverse lookup keys.
func (r *ForwardLookupOnlyReconciler) Reconcile(ctx context.Context) (int, error) {
	return r.reconcileForwardRedirects(ctx)
}

func (r *ForwardLookupOnlyReconciler) reconcileForwardRedirects(ctx context.Context) (int, error) {
	return reconcileBatch(ctx, r.reconcileBatchForForwardLookupOnly)
}

func (r *ForwardLookupOnlyReconciler) reconcileBatchForForwardLookupOnly(ctx context.Context, cursor uint64) (recordsChanged int, newCursor uint64, err error) {
	fwdRedirects, newCursor, err := r.ds.GetKeys(ctx, fmt.Sprintf("%s*", fwdPrefix), 100, cursor)
	if err != nil {
		return recordsChanged, newCursor, fmt.Errorf("get redirects with match failed: %w", err)
	}

	for _, key := range fwdRedirects {
		err := r.ds.ConvertRedirectFormatToForwardLookupOnly(ctx, key)
		if err != nil {
			return recordsChanged, newCursor, fmt.Errorf("convert redirect format to forward lookup only failed: %w", err)
		}
		recordsChanged++
	}
	return recordsChanged, newCursor, nil
}

func reconcileBatch(ctx context.Context, reconcileFunc func(ctx context.Context, cursor uint64) (recordsChanged int, newCursor uint64, err error)) (int, error) {
	var recordsChanged int
	var cursor uint64
	var err error
	for {
		var recordsChangedBatch int
		recordsChangedBatch, cursor, err = reconcileFunc(ctx, cursor)
		if err != nil {
			return recordsChanged, fmt.Errorf("reconcile batch scan failed: cursor=%d: %w", cursor, err)
		}
		recordsChanged += recordsChangedBatch
		if cursor == 0 {
			break
		}
	}
	return recordsChanged, nil
}
