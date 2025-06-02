package service_test

import (
	"errors"
	"strings"
	"testing"

	"vidra-cli/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestGetClusterKubeConfigSecret(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	cs := service.NewClusterService(mockCLI)

	mockCLI.On("GetByLabel", mock.Anything, "secret", "test-ns", "cluster-kubeconfig", "test-cluster").
		Return([]byte("some-secret"), nil)

	err := cs.PrintClusterKubeConfigSecret("test-cluster", "test-ns")
	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestRemoveClusterKubeConfigSecret(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	cs := service.NewClusterService(mockCLI)

	mockCLI.On("Hash", "test-cluster").Return("abc123")
	mockCLI.On("Delete", mock.Anything, "secret", "cluster-kubeconfig-abc123", "test-ns").
		Return(nil)

	err := cs.RemoveClusterKubeConfigSecret("test-cluster", "test-ns")
	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestListClusterKubeConfigSecrets(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	cs := service.NewClusterService(mockCLI)

	mockCLI.On("ListByLabel", mock.Anything, "secrets", "cluster-kubeconfig",
		"custom-columns=SECRET-NAME:.metadata.name,NAMESPACE:.metadata.namespace,CLUSTER_NAME:.metadata.labels.cluster-kubeconfig").
		Return([]byte("list-result"), nil)

	err := cs.ListClusterKubeConfigSecrets()
	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}
func TestApplyClusterKubeConfigSecret_Success(t *testing.T) {
	mockKube := new(mockKubeCLI)
	svc := service.NewClusterService(mockKube)

	// Build a fake kubeconfig
	fakeConfig := api.NewConfig()
	fakeConfig.Clusters["cluster1"] = &api.Cluster{Server: "http://server"}
	fakeConfig.AuthInfos["auth1"] = &api.AuthInfo{}
	fakeConfig.Contexts["my-cluster"] = &api.Context{
		Cluster:  "cluster1",
		AuthInfo: "auth1",
	}

	loader := func(path string) (*api.Config, error) {
		if path != "/fake/path" {
			t.Fatalf("unexpected path: %s", path)
		}
		return fakeConfig, nil
	}

	// clientcmd.Write serializes the config, we can just test that string passed to EncodeBase64 is non-empty
	mockKube.On("EncodeBase64", mock.Anything).Return("encoded-kubeconfig").Once()
	mockKube.On("LabelFromURL", "http://server").Return("label-for-server", nil).Once()
	mockKube.On("Hash", "my-cluster").Return("hash-my-cluster").Once()
	mockKube.On("ApplyYAML", mock.Anything, mock.MatchedBy(func(yaml string) bool {
		return strings.Contains(yaml, "name: cluster-kubeconfig-hash-my-cluster") &&
			strings.Contains(yaml, "namespace: myns") &&
			strings.Contains(yaml, "cluster-kubeconfig: label-for-server") &&
			strings.Contains(yaml, "kubeconfig: encoded-kubeconfig")
	})).Return(nil).Once()

	err := svc.ApplyClusterKubeConfigSecret("my-cluster", "myns", "/fake/path", loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockKube.AssertExpectations(t)
}

func TestApplyClusterKubeConfigSecret_LoadError(t *testing.T) {
	mockKube := new(mockKubeCLI)
	svc := service.NewClusterService(mockKube)

	loader := func(path string) (*api.Config, error) {
		return nil, errors.New("load failure")
	}

	err := svc.ApplyClusterKubeConfigSecret("any-cluster", "ns", "/fake/path", loader)
	if err == nil || !strings.Contains(err.Error(), "failed to load kubeconfig") {
		t.Fatalf("expected load kubeconfig error, got %v", err)
	}
}

func TestApplyClusterKubeConfigSecret_ContextNotFound(t *testing.T) {
	mockKube := new(mockKubeCLI)
	svc := service.NewClusterService(mockKube)

	fakeConfig := api.NewConfig() // no contexts
	loader := func(path string) (*api.Config, error) {
		return fakeConfig, nil
	}

	err := svc.ApplyClusterKubeConfigSecret("missing-context", "ns", "/fake/path", loader)
	if err == nil || !strings.Contains(err.Error(), `context "missing-context" not found`) {
		t.Fatalf("expected context not found error, got %v", err)
	}
}

func TestApplyClusterKubeConfigSecret_ClusterNotFound(t *testing.T) {
	mockKube := new(mockKubeCLI)
	svc := service.NewClusterService(mockKube)

	fakeConfig := api.NewConfig()
	fakeConfig.Contexts["ctx1"] = &api.Context{Cluster: "missing-cluster", AuthInfo: "auth"}
	loader := func(path string) (*api.Config, error) {
		return fakeConfig, nil
	}

	err := svc.ApplyClusterKubeConfigSecret("ctx1", "ns", "/fake/path", loader)
	if err == nil || !strings.Contains(err.Error(), `cluster "missing-cluster" not found`) {
		t.Fatalf("expected cluster not found error, got %v", err)
	}
}

func TestApplyClusterKubeConfigSecret_LabelFromURLError(t *testing.T) {
	mockKube := new(mockKubeCLI)
	svc := service.NewClusterService(mockKube)

	fakeConfig := api.NewConfig()
	fakeConfig.Clusters["cluster1"] = &api.Cluster{Server: "bad-url"}
	fakeConfig.AuthInfos["auth1"] = &api.AuthInfo{}
	fakeConfig.Contexts["ctx1"] = &api.Context{
		Cluster:  "cluster1",
		AuthInfo: "auth1",
	}

	loader := func(path string) (*api.Config, error) {
		return fakeConfig, nil
	}

	mockKube.On("EncodeBase64", mock.Anything).Return("encoded").Once()
	mockKube.On("LabelFromURL", "bad-url").Return("", errors.New("invalid URL")).Once()

	err := svc.ApplyClusterKubeConfigSecret("ctx1", "ns", "/fake/path", loader)
	if err == nil || !strings.Contains(err.Error(), "invalid server URL") {
		t.Fatalf("expected invalid server URL error, got %v", err)
	}

	mockKube.AssertExpectations(t)
}
