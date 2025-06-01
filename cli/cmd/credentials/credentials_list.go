package credentials

import (
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Infrahub credentials secrets",
	Run: func(cmd *cobra.Command, args []string) {
		credentialsService := setup()
		err := credentialsService.ListCredentialsSecrets()
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}

	},
}
