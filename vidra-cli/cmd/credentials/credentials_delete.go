package credentials

import (
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <url>",
	Short: "Remove a Infrahub credential secret by URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		urlStr := args[0]
		credentialsService := setup()
		err := credentialsService.RemoveCredentialsSecret(urlStr, namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"vidra-system\")")
}
