package k8s

import (
	"context"
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Use label selector to find the Secret based on the InfrahubAPIURL label
func GetSortedListByLabel(
	ctx context.Context,
	k8sClient client.Client,
	labelKey, labelValue string,
	list client.ObjectList,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Getting sorted list by label", "labelKey", labelKey, "labelValue", labelValue)
	if err := k8sClient.List(ctx, list, client.MatchingLabels{labelKey: labelValue}); err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	items, err := meta.ExtractList(list)
	if err != nil {
		return fmt.Errorf("failed to extract items: %w", err)
	}

	if items == nil {
		return fmt.Errorf("extracted list is nil for label %s=%s", labelKey, labelValue)
	}
	if len(items) == 0 {
		return fmt.Errorf("no resources found with label %s=%s", labelKey, labelValue)
	}

	// Sort items by creation timestamp descending
	sort.Slice(items, func(i, j int) bool {
		iMeta, _ := meta.Accessor(items[i])
		jMeta, _ := meta.Accessor(items[j])
		return iMeta.GetCreationTimestamp().After(jMeta.GetCreationTimestamp().Time)
	})

	// Set sorted items back into the list
	if err := meta.SetList(list, items); err != nil {
		return fmt.Errorf("failed to set sorted items into list: %w", err)
	}

	return nil
}
