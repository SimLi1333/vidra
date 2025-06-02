package service

import (
	"context"
	"fmt"

	"github.com/infrahub-operator/vidra/vidra-cli/internal/adapter/kubecli"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type clusterService struct {
	kubecli kubecli.KubeCLI
}

func NewClusterService(cli kubecli.KubeCLI) ClusterService {
	return &clusterService{kubecli: cli}
}

func (s *clusterService) PrintClusterKubeConfigSecret(clusterName, namespace string) error {
	yaml, err := s.kubecli.GetByLabel(context.Background(), "secret", namespace, "cluster-kubeconfig", clusterName)
	if err != nil {
		return err
	}
	fmt.Println(string(yaml) + "\n\n---\n")
	return nil
}

type KubeConfigLoader func(path string) (*api.Config, error)

func (s *clusterService) ApplyClusterKubeConfigSecret(
	clusterName, namespace, kubeconfigPath string,
	loadKubeConfig KubeConfigLoader,
) error {
	print(kubeconfigPath)
	config, err := loadKubeConfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	fmt.Println("Contexts loaded from config:")
	for ctx := range config.Contexts {
		fmt.Println(" -", ctx)
	}

	ctxCfg, exists := config.Contexts[clusterName]
	if !exists {
		return fmt.Errorf("context %q not found in kubeconfig", clusterName)
	}

	cluster := config.Clusters[ctxCfg.Cluster]
	if cluster == nil {
		return fmt.Errorf("cluster %q not found in kubeconfig", ctxCfg.Cluster)
	}

	newConfig := api.NewConfig()
	newConfig.CurrentContext = clusterName
	newConfig.Contexts[clusterName] = ctxCfg
	newConfig.Clusters[ctxCfg.Cluster] = cluster
	newConfig.AuthInfos[ctxCfg.AuthInfo] = config.AuthInfos[ctxCfg.AuthInfo]

	kubeconfigData, err := clientcmd.Write(*newConfig)
	if err != nil {
		return fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	encoded := s.kubecli.EncodeBase64(string(kubeconfigData))

	label, err := s.kubecli.LabelFromURL(cluster.Server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	secretYAML := generateClusterSecretYAML(s.kubecli.Hash(clusterName), label, encoded, namespace)
	fmt.Println(secretYAML + "\n\n---\n")
	fmt.Println()
	return s.kubecli.ApplyYAML(context.Background(), secretYAML)
}

func (s *clusterService) ListClusterKubeConfigSecrets() error {
	result, err := s.kubecli.ListByLabel(
		context.Background(),
		"secrets",
		"cluster-kubeconfig",
		"custom-columns=SECRET-NAME:.metadata.name,NAMESPACE:.metadata.namespace,CLUSTER_NAME:.metadata.labels.cluster-kubeconfig",
	)
	if err != nil {
		return err
	}
	fmt.Println(string(result) + "\n\n---\n")
	return nil
}

func (s *clusterService) RemoveClusterKubeConfigSecret(clusterName, namespace string) error {
	secretName := fmt.Sprintf("cluster-kubeconfig-%s", s.kubecli.Hash(clusterName))
	return s.kubecli.Delete(context.Background(), "secret", secretName, namespace)
}

func generateClusterSecretYAML(clusterID, clusterLabel, encoded, namespace string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: cluster-kubeconfig-%s
  namespace: %s
  labels:
    cluster-kubeconfig: %s
type: Opaque
data:
  kubeconfig: %s
`, clusterID, namespace, clusterLabel, encoded)
}
