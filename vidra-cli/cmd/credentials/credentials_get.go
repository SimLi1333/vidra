package credentials

import (
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <url>",
	Short: "Get a Infrahub credentials secret by URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		credentialsService := setup()
		err := credentialsService.PrintCredentialsSecret(args[0], namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	getCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"vidra-system\")")
}
