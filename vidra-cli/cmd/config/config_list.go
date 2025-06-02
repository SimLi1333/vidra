package config

import (
	"os"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all vidra-operator config ConfigMaps in the namespace",
	Run: func(cmd *cobra.Command, args []string) {
		configService := setup()
		err := configService.ListConfigMaps()
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}
