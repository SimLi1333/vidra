package service_test

import (
	"errors"
	"testing"
	"vidra-cli/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPrintInfrahubSync(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	svc := service.NewInfrahubSyncService(mockCLI)

	mockCLI.On("LabelFromURL", "http://some-url").Return("some-url", nil)
	mockCLI.On("GetByName", mock.Anything, "infrahubSync", "default", "mysync").Return([]byte("test-yaml"), nil)

	err := svc.PrintInfrahubSync("http://some-url", "default", "mysync")
	assert.NoError(t, err)

	mockCLI.AssertExpectations(t)
}

func TestApplyInfrahubSync(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	svc := service.NewInfrahubSyncService(mockCLI)

	mockCLI.On("LabelFromURL", "http://some-url").Return("some-url", nil)
	mockCLI.On("ApplyYAML", mock.Anything, mock.MatchedBy(func(s string) bool {
		return len(s) > 0 && s[0:10] == "apiVersion"
	})).Return(nil)

	err := svc.ApplyInfrahubSync(
		"http://some-url", "artifact", "main", "2024-01-01",
		"my-server", "dest-ns", true, "default", "mysync",
	)
	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestListInfrahubSync(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	svc := service.NewInfrahubSyncService(mockCLI)

	expectedOutput := []byte("name\tnamespace\t...")
	mockCLI.On("ListByLabel", mock.Anything, "infrahubsyncs", "", mock.Anything).Return(expectedOutput, nil)

	err := svc.ListInfrahubSync()
	assert.NoError(t, err)

	mockCLI.AssertExpectations(t)
}

func TestRemoveInfrahubSync(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	svc := service.NewInfrahubSyncService(mockCLI)

	mockCLI.On("LabelFromURL", "http://some-url").Return("some-url", nil)
	mockCLI.On("Delete", mock.Anything, "infrahubSync", "mysync", "default").Return(nil)

	err := svc.RemoveInfrahubSync("http://some-url", "default", "mysync")
	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestRemoveInfrahubSync_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "https://infrahub.example.com").Return("infrahub.example.com", nil)
	mockCLI.On("Delete", mock.Anything, "infrahubSync", "mysync", "myns").Return(nil)

	svc := service.NewInfrahubSyncService(mockCLI)

	err := svc.RemoveInfrahubSync("https://infrahub.example.com", "myns", "mysync")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestRemoveInfrahubSync_LabelError(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "bad-url").Return("bad-url", errors.New("invalid url"))
	mockCLI.On("Delete", mock.Anything, "infrahubSync", "mysync", "myns").Return(nil)

	svc := service.NewInfrahubSyncService(mockCLI)

	err := svc.RemoveInfrahubSync("bad-url", "myns", "mysync")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}
