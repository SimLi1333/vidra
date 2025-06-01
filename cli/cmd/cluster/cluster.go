package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"vidra-cli/internal/adapter/kubecli"
	"vidra-cli/internal/service"

	"github.com/spf13/cobra"
)

var (
	clusterName string
	namespace   string
)

var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters for multticluster vidra-operator",
}

func init() {
	ClusterCmd.AddCommand(applyCmd)
	ClusterCmd.AddCommand(getCmd)
	ClusterCmd.AddCommand(listCmd)
	ClusterCmd.AddCommand(deleteCmd)
}

func setup() service.ClusterService {
	cli := kubecli.NewDefaultKubeCLI()
	return service.NewClusterService(cli)
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
