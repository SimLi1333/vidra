package credentials

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"vidra-cli/internal/adapter/kubecli"
	"vidra-cli/internal/service"

	"github.com/spf13/cobra"
)

var namespace string

var CredentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage Infrahub credential Secrets",
}

func init() {

	CredentialsCmd.AddCommand(applyCmd)
	CredentialsCmd.AddCommand(getCmd)
	CredentialsCmd.AddCommand(listCmd)
	CredentialsCmd.AddCommand(deleteCmd)
}

var setupFn = setup

func setup() service.CredentialsService {
	cli := kubecli.NewDefaultKubeCLI()
	return service.NewCredentialsService(cli)
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
