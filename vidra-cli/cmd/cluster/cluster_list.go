package cluster

import (
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cluster kubeconfig Secrets for the vidra-operator",
	Run: func(cmd *cobra.Command, args []string) {
		clusterService := setup()
		err := clusterService.ListClusterKubeConfigSecrets()
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}
