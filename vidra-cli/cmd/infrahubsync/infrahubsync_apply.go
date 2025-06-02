package infrahubsync

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

var (
	server           string
	destNamespace    string
	reconcileOnEvent bool
)

var applyCmd = &cobra.Command{
	Use:   "apply <url>",
	Short: "Generate InfrahubSync and apply it to Kubernetes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		if date != "" {
			if !IsValidTargetDateFormat(date) {
				fmt.Fprintf(os.Stderr, "Error: Invalid targetDate format. Must be RFC3339 or relative like 'now-2h', got: %s\n", date)
				os.Exit(1)
			}
		}
		if artifact == "" {
			fmt.Fprintln(os.Stderr, "Error: artifactName is required")
			cmd.Usage()
			os.Exit(1)
		}
		infrahubSyncService := setup()
		if name == "" {
			name = "infrahubsync-" + generateHash(url, date, branche, artifact)
		}
		err := infrahubSyncService.ApplyInfrahubSync(url, artifact, branche, date, server, destNamespace, reconcileOnEvent, namespace, name)
		if err != nil {
			errorHandler(err)
			os.Exit(1)
		}
	},
}

func init() {
	applyCmd.Flags().StringVarP(&namespace, "InfrahubSync namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"vidra-system\")")
	applyCmd.Flags().StringVarP(&branche, "targetBranche", "b", "main", "Infrahub branche to sync to")
	applyCmd.Flags().StringVarP(&date, "targetDate", "d", "", "Date and time to sync with Infrahub (RFC3339 format) or relative format (e.g. 5m, 2h))")
	applyCmd.Flags().StringVarP(&artifact, "artifactName", "a", "", "Name of the artifact definition in Infrahub to sync to")
	applyCmd.MarkFlagRequired("artifact")
	applyCmd.Flags().StringVarP(&server, "server", "s", "", "Destination Kubernetes server (optional, defaults to local cluster)")
	applyCmd.Flags().BoolVarP(&reconcileOnEvent, "reconcileOnEvent", "e", false, "if -e added reconcile on Infrahub events (default: false)")
	applyCmd.Flags().StringVarP(&destNamespace, "destinationNamespace", "N", "", "Destination Kubernetes namespace (optional, defaults to the default namespace)")
	applyCmd.Flags().StringVarP(&name, "InfrahubSync name", "E", "", "Name of the InfrahubSync resource (optional, defaults to a generated name based on the URL)")
}

var relativeFormatRegex = regexp.MustCompile(`^[a-zA-Z]+[-+]\d+[smh]$`)

// IsValidTargetDateFormat checks if the input string is a valid RFC3339 or relative time format.
func IsValidTargetDateFormat(input string) bool {
	if _, err := time.Parse(time.RFC3339, input); err == nil {
		return true
	}
	if relativeFormatRegex.MatchString(input) {
		return true
	}
	return false
}
