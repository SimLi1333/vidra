package infrahubsync

import (
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <url>",
	Short: "Get a InfrahubSync by URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		infrahubSyncService := setup()
		if name == "" {
			name = "infrahubsync-" + generateHash(url, date, branche, artifact)
		}
		err := infrahubSyncService.PrintInfrahubSync(args[0], namespace, name)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	getCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"vidra-system\")")
	getCmd.Flags().StringVarP(&name, "InfrahubSync name", "N", "", "Name of the InfrahubSync resource (optional, defaults to a generated name based on the URL)")
	getCmd.Flags().StringVarP(&branche, "targetBranche", "b", "main", "Infrahub branche to sync to")
	getCmd.Flags().StringVarP(&date, "targetDate", "d", "", "Date and time to sync with Infrahub (RFC3339 format) or relative format (e.g. 5m, 2h))")
	getCmd.Flags().StringVarP(&artifact, "artifactName", "a", "", "Name of the artifact definition in Infrahub to sync to")

}
