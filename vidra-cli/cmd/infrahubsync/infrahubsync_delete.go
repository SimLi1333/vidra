package infrahubsync

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
		infrahubSyncService := setup()
		if name == "" {
			name = "infrahubsync-" + generateHash(urlStr, date, branche, artifact)
		}
		err := infrahubSyncService.RemoveInfrahubSync(urlStr, namespace, name)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&namespace, "InfrahubSync namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"vidra-system\")")
	deleteCmd.Flags().StringVarP(&name, "InfrahubSync name", "N", "", "Name of the InfrahubSync resource (optional, defaults to a generated name based on the URL)")
	deleteCmd.Flags().StringVarP(&branche, "targetBranche", "b", "main", "Infrahub branche to sync to")
	deleteCmd.Flags().StringVarP(&date, "targetDate", "d", "", "Date and time to sync with Infrahub (RFC3339 format) or relative format (e.g. 5m, 2h))")
	deleteCmd.Flags().StringVarP(&artifact, "artifactName", "a", "", "Name of the artifact definition in Infrahub to sync to")
}
