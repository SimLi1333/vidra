package config

import (
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch the vidra-operator config ConfigMap",
	Run: func(cmd *cobra.Command, args []string) {
		configService := setup()
		err := configService.PrintConfigMap(namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	getCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the configMap (default: \"default\")")
}
