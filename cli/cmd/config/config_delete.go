package config

import (
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the vidra-operator config ConfigMap",
	Run: func(cmd *cobra.Command, args []string) {
		configService := setup()
		err := configService.RemoveConfigMap(namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the configMap (default: \"default\")")
}
