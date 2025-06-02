package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/infrahub-operator/vidra/vidra-cli/internal/adapter/kubecli"
	"github.com/infrahub-operator/vidra/vidra-cli/internal/service"

	"github.com/spf13/cobra"
)

var namespace string

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "configure the Vidra Operator",
}

func init() {
	ConfigCmd.AddCommand(applyCmd)
	ConfigCmd.AddCommand(getCmd)
	ConfigCmd.AddCommand(listCmd)
	ConfigCmd.AddCommand(deleteCmd)
}

func setup() service.ConfigService {
	cli := kubecli.NewDefaultKubeCLI()
	return service.NewConfigService(cli)
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
