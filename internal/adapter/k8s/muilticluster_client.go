package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DynamicClientFactory struct {
	mu      sync.Mutex
	clients map[string]client.Client
}

func NewDynamicClientFactory() *DynamicClientFactory {
	return &DynamicClientFactory{
		clients: make(map[string]client.Client),
	}
}

func (f *DynamicClientFactory) GetCachedClientFor(ctx context.Context, serverURL string, k8sClient client.Client) (client.Client, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if cached, ok := f.clients[serverURL]; ok {
		return cached, nil
	}

	secretList := &v1.SecretList{}

	trimmedK8SURL := strings.TrimPrefix(strings.Split(serverURL, ":")[1], "//")
	err := GetSortedListByLabel(ctx, k8sClient, "cluster-kubeconfig", trimmedK8SURL, secretList)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets by label: %w", err)
	}

	var kubeConfigData []byte
	for _, secret := range secretList.Items {
		if data, exists := secret.Data["kubeconfig"]; exists {
			kubeConfigData = data
			break
		}
	}
	if kubeConfigData == nil {
		return nil, fmt.Errorf("kubeconfig not found in any secret")
	}

	rawConfig, err := clientcmd.Load(kubeConfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig data: %w", err)
	}

	// Select matching context dynamically
	var selectedContext string
	for name, ctx := range rawConfig.Contexts {
		cluster := rawConfig.Clusters[ctx.Cluster]
		if cluster != nil && strings.Contains(cluster.Server, trimmedK8SURL) {
			selectedContext = name
			break
		}
	}
	if selectedContext == "" {
		return nil, fmt.Errorf("no matching context found and current context is empty")
	}

	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: selectedContext,
	}

	clientConfig := clientcmd.NewNonInteractiveClientConfig(*rawConfig, selectedContext, configOverrides, nil)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config from kubeconfig: %w", err)
	}

	cachedClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create cached client: %w", err)
	}

	f.clients[serverURL] = cachedClient
	return cachedClient, nil
}
