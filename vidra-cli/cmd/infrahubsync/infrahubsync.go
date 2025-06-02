package infrahubsync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"vidra-cli/internal/adapter/kubecli"
	"vidra-cli/internal/service"

	"github.com/spf13/cobra"
)

var (
	namespace string
	name      string
	date      string
	branche   string
	artifact  string
)

var InfrahubSyncCmd = &cobra.Command{
	Use:   "infrahubsync",
	Short: "Manage InfrahubSync resources",
}

func init() {

	InfrahubSyncCmd.AddCommand(applyCmd)
	InfrahubSyncCmd.AddCommand(getCmd)
	InfrahubSyncCmd.AddCommand(listCmd)
	InfrahubSyncCmd.AddCommand(deleteCmd)
}

var setupFn = setup

func setup() service.InfrahubSyncService {
	cli := kubecli.NewDefaultKubeCLI()
	return service.NewInfrahubSyncService(cli)
}

func errorHandler(err error) {
	if strings.Contains(err.Error(), "signal: killed") {
		fmt.Fprintln(os.Stderr, "Error: operation timed out.")
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		fmt.Fprintf(os.Stderr, "stderr: %s\n", exitErr.Stderr)
	}
}

// generateHash creates a SHA-256 hash from the given url, date, branch, and artifact.
func generateHash(url, date, branche, artifact string) string {
	data := url + date + branche + artifact
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
