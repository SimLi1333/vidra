package main

import (
	"fmt"
	"os"

	"github.com/infrahub-operator/vidra/vidra-cli/cmd/cluster" // Import cmd package
	"github.com/infrahub-operator/vidra/vidra-cli/cmd/config"  // Import cmd package
	"github.com/infrahub-operator/vidra/vidra-cli/cmd/credentials"
	"github.com/infrahub-operator/vidra/vidra-cli/cmd/infrahubsync"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vidra",
	Short: "CLI to manage Infrahub secrets",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Register the repoCmd from cmd package to rootCmd
	rootCmd.AddCommand(credentials.CredentialsCmd)
	rootCmd.AddCommand(cluster.ClusterCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(infrahubsync.InfrahubSyncCmd)
}
