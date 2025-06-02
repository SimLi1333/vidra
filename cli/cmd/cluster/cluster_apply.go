package cluster

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var applyCmd = &cobra.Command{
	Use:   "apply <context>",
	Short: "Apply the cluster kubeconfig secret for the vidra-operator",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: accepts 1 arg(s), received 0")
			cmd.Usage()
			printCurrentKubeContext()
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Error: accepts 1 arg(s), received 2 or more")
			cmd.Usage()
			printCurrentKubeContext()
			os.Exit(1)
		}
		kubeContext := args[0]
		clusterService := setup()
		err := clusterService.ApplyClusterKubeConfigSecret(kubeContext, namespace, clientcmd.RecommendedHomeFile,
			defaultKubeConfigLoader)
		if err != nil {
			errorHandler(err)
			printCurrentKubeContext()
			cmd.Usage()
			os.Exit(1)
		}
	},
}

func init() {
	applyCmd.Flags().StringVarP(&namespace, "namespace", "n", "vidra-system", "Kubernetes namespace for the secret (default: \"default\")")
}

func printCurrentKubeContext() {
	fmt.Println("\nCurrent Kubernetes context:")
	cmd := exec.Command("kubectl", "config", "get-contexts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get current context: %v\n", err)
	}
}

func defaultKubeConfigLoader(path string) (*api.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules() // respects $KUBECONFIG or defaults to ~/.kube/config
	configOverrides := &clientcmd.ConfigOverrides{}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}

	return &rawConfig, nil
}
