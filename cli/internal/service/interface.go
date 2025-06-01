package service

type CredentialsService interface {
	ApplyCredentialsSecret(url, username, password, namespace string) error
	PrintCredentialsSecret(url, namespace string) error
	ListCredentialsSecrets() error
	RemoveCredentialsSecret(url, namespace string) error
}

type ConfigService interface {
	ApplyConfigMap(requeueSyncAfter, requeueResourceAfter, queryName, namespace string) error
	PrintConfigMap(namespace string) error
	ListConfigMaps() error
	RemoveConfigMap(namespace string) error
}

type ClusterService interface {
	ApplyClusterKubeConfigSecret(clusterName, namespace, kubeconfigPath string, loadKubeConfig KubeConfigLoader) error
	PrintClusterKubeConfigSecret(clusterName, namespace string) error
	ListClusterKubeConfigSecrets() error
	RemoveClusterKubeConfigSecret(clusterName, namespace string) error
}
