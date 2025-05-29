package controller

import (
	"context"
	"fmt"

	infrahubv1alpha1 "github.com/infrahub-operator/vidra/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Sync and Resource state helpers

// MarkSyncState updates the resource's SyncState using the provided update function.
// It retries on conflict using the default backoff strategy.
func MarkState(
	ctx context.Context,
	c client.StatusClient,
	res client.Object,
	updateStatus func(),
) error {
	if res == nil {
		return fmt.Errorf("resource is nil")
	}

	// Create a patch based on the original resource state
	patch := client.MergeFrom(res.DeepCopyObject().(client.Object))
	// Apply the status updates
	updateStatus()

	// Retry patching on conflict
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return c.Status().Patch(ctx, res, patch)
	}); err != nil {
		return fmt.Errorf("failed to patch SyncState: %w", err)
	}
	return nil
}

// MarkSyncFailed sets the resource's SyncState to Failed and logs the error.
// It calls MarkSyncState to handle the update, retrying on conflict.
func MarkStateFailed(
	ctx context.Context,
	c client.StatusClient,
	res client.Object,
	originalErr error,
) error {
	logger := log.FromContext(ctx)
	if res == nil {
		return fmt.Errorf("resource is nil")
	}
	// Update the LastError field and SyncState to Failed
	err := MarkState(ctx, c, res, func() {
		switch obj := res.(type) {
		case *infrahubv1alpha1.VidraResource:
			obj.Status.LastSyncTime = metav1.Now()
			obj.Status.LastError = originalErr.Error()
			obj.Status.DeployState = infrahubv1alpha1.StateFailed
		case *infrahubv1alpha1.InfrahubSync:
			obj.Status.LastSyncTime = metav1.Now()
			obj.Status.LastError = originalErr.Error()
			obj.Status.SyncState = infrahubv1alpha1.StateFailed
		default:
			// Log unsupported resource type error
			logger.Error(fmt.Errorf("unsupported resource type"), "failed to update resource status")
		}
	})

	// Log if there was an error updating the SyncState to Failed
	if err != nil {
		logger.Error(err, "failed to update SyncState to Failed")
	}

	return originalErr
}
