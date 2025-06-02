package infrahubsync

import (
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all InfrahubSync",
	Run: func(cmd *cobra.Command, args []string) {
		infrahubSyncService := setup()
		err := infrahubSyncService.ListInfrahubSync()
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}
