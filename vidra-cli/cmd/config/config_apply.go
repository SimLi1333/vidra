package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	requeueSyncAfter     string
	requeueResourceAfter string
	queryName            string
	eventBasedReconcile  bool
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Generate and apply the vidra-operator config ConfigMap",
	Run: func(cmd *cobra.Command, args []string) {
		configService := setup()
		if !IsValidDurationFormat(requeueSyncAfter) {
			errorHandler(fmt.Errorf("invalid requeue-sync-after format: %s", requeueSyncAfter))
			os.Exit(1)
		}
		if !IsValidDurationFormat(requeueResourceAfter) {
			errorHandler(fmt.Errorf("invalid requeue-resource-after format: %s", requeueResourceAfter))
			os.Exit(1)
		}
		err := configService.ApplyConfigMap(requeueSyncAfter, requeueResourceAfter, queryName, eventBasedReconcile, namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	applyCmd.Flags().StringVarP(&requeueSyncAfter, "requeue-sync-after", "s", "1m", "Requeue duration of infrahub Sync (e.g. 30s, 5m, 2h)")
	applyCmd.Flags().StringVarP(&requeueResourceAfter, "requeue-resource-after", "r", "1m", "Requeue duration of k8 reconciliation (e.g. 30s, 5m, 2h)")
	applyCmd.Flags().StringVarP(&queryName, "query-name", "q", "ArtifactIDs", "Name of the Infrahub query")
	applyCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the configMap (default: \"default\")")
	applyCmd.Flags().BoolVarP(&eventBasedReconcile, "eventBasedReconcile", "e", false, "Enable global event-based reconciliation for vidra (default: false)")
}

func IsValidDurationFormat(input string) bool {
	_, err := time.ParseDuration(input)
	return err == nil
}
