package credentials

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	username string
	password string
)

var applyCmd = &cobra.Command{
	Use:   "apply <url>",
	Short: "Generate and apply Infrahub credentials secret to Kubernetes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		if username == "" || password == "" {
			fmt.Println("Both --username and --password are required")
			cmd.Usage()
			return
		}
		credentialsService := setup()
		err := credentialsService.ApplyCredentialsSecret(url, username, password, namespace)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	applyCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"vidra-system\")")
	applyCmd.Flags().StringVarP(&username, "username", "u", "", "Infrahub username (required)")
	applyCmd.Flags().StringVarP(&password, "password", "p", "", "Infrahub password (required)")
	applyCmd.MarkFlagRequired("username")
	applyCmd.MarkFlagRequired("password")
}
