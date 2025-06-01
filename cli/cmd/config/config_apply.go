package config

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	requeueSyncAfter     string
	requeueResourceAfter string
	queryName            string
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Generate and apply the vidra-operator config ConfigMap",
	Run: func(cmd *cobra.Command, args []string) {
		configService := setup()
		err := configService.ApplyConfigMap(requeueSyncAfter, requeueResourceAfter, queryName, namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	applyCmd.Flags().StringVarP(&requeueSyncAfter, "requeue-sync-after", "s", "1m", "Requeue duration of infrahub Sync (e.g. 30s, 5m, 2h)")
	applyCmd.Flags().StringVarP(&requeueSyncAfter, "requeue-resource-after", "r", "1m", "Requeue duration of k8 reconciliation (e.g. 30s, 5m, 2h)")
	applyCmd.Flags().StringVarP(&queryName, "query-name", "q", "ArtifactIDs", "Name of the Infrahub query")
	applyCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the configMap (default: \"default\")")
}
