package cluster

import (
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch the cluster kubeconfig secret for the vidra-operator",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName = args[0]
		clusterService := setup()
		err := clusterService.PrintClusterKubeConfigSecret(clusterName, namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	getCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"default\")")
}
