package service_test

import (
	"errors"
	"testing"

	"vidra-cli/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfigMap_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("GetByLabel", mock.Anything, "configmap", "default", "app", "vidra").
		Return([]byte("config-yaml"), nil)

	svc := service.NewConfigService(mockCLI)

	err := svc.PrintConfigMap("default")
	assert.NoError(t, err)

	mockCLI.AssertExpectations(t)
}

func TestGetConfigMap_Error(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("GetByLabel", mock.Anything, "configmap", "default", "app", "vidra").
		Return([]byte(nil), errors.New("not found"))

	svc := service.NewConfigService(mockCLI)

	err := svc.PrintConfigMap("default")
	assert.EqualError(t, err, "not found")

	mockCLI.AssertExpectations(t)
}

func TestApplyConfigMap(t *testing.T) {
	mockCLI := new(mockKubeCLI)

	expectedYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: vidra-config
  namespace: default
  labels:
    app: vidra
data:
  requeueSyncAfter: "15m"
  requeueResourceAfter: "20m"
  queryName: "sync-artifacts"
`
	mockCLI.On("ApplyYAML", mock.Anything, expectedYAML).Return(nil)

	svc := service.NewConfigService(mockCLI)
	err := svc.ApplyConfigMap("15m", "20m", "sync-artifacts", "default")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestListConfigMaps_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)

	mockCLI.On("ListByLabel", mock.Anything,
		"configmaps",
		"app=vidra",
		"custom-columns=CONFIGMAP-NAME:.metadata.name,NAMESPACE:.metadata.namespace,REQUEUE_SYNC_AFTER:.data.requeueSyncAfter,REQUEUE_RESOURCE_AFTER:.data.requeueResourceAfter,QUERY_NAME:.data.queryName").
		Return([]byte("configmap-list"), nil)

	svc := service.NewConfigService(mockCLI)

	err := svc.ListConfigMaps()
	assert.NoError(t, err)

	mockCLI.AssertExpectations(t)
}

func TestListConfigMaps_Error(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("ListByLabel", mock.Anything, "configmaps", "app=vidra", mock.Anything).
		Return([]byte(nil), errors.New("list failed"))

	svc := service.NewConfigService(mockCLI)

	err := svc.ListConfigMaps()
	assert.EqualError(t, err, "list failed")

	mockCLI.AssertExpectations(t)
}

func TestRemoveConfigMap(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("Delete", mock.Anything, "configmap", "vidra-config", "default").Return(nil)

	svc := service.NewConfigService(mockCLI)

	err := svc.RemoveConfigMap("default")
	assert.NoError(t, err)

	mockCLI.AssertExpectations(t)
}
