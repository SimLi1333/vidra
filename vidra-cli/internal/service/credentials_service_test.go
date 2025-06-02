package service_test

import (
	"errors"
	"strings"
	"testing"
	"vidra-cli/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCredentialsSecret_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "https://infrahub.example.com").Return("infrahub.example.com", nil)
	mockCLI.On("GetByLabel", mock.Anything, "secret", "test-ns", "infrahub-api-url", "infrahub.example.com").
		Return([]byte("secret-yaml"), nil)

	svc := service.NewCredentialsService(mockCLI)
	err := svc.PrintCredentialsSecret("https://infrahub.example.com", "test-ns")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestGetCredentialsSecret_LabelFromURL_Error(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "bad-url").Return("bad-url", errors.New("invalid url"))
	mockCLI.On("GetByLabel", mock.Anything, "secret", "ns", "infrahub-api-url", "bad-url").
		Return([]byte("fallback"), nil)

	svc := service.NewCredentialsService(mockCLI)
	err := svc.PrintCredentialsSecret("bad-url", "ns")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestApplyCredentialsSecret_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "https://infrahub.example.com").Return("infrahub.example.com", nil)
	mockCLI.On("Hash", "infrahub.example.com").Return("abcd1234")
	mockCLI.On("EncodeBase64", "https://infrahub.example.com").Return("base64-url")
	mockCLI.On("EncodeBase64", "user").Return("base64-user")
	mockCLI.On("EncodeBase64", "pass").Return("base64-pass")
	mockCLI.On("ApplyYAML", mock.Anything, mock.MatchedBy(func(yaml string) bool {
		return strings.Contains(yaml, "infrahub-credentials-abcd1234") &&
			strings.Contains(yaml, "base64-user") &&
			strings.Contains(yaml, "base64-pass") &&
			strings.Contains(yaml, "base64-url")
	})).Return(nil)

	svc := service.NewCredentialsService(mockCLI)
	err := svc.ApplyCredentialsSecret("https://infrahub.example.com", "user", "pass", "my-ns")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestListCredentialsSecrets_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("ListByLabel", mock.Anything,
		"secrets",
		"infrahub-api-url",
		"custom-columns=SECRET-NAME:.metadata.name,NAMESPACE:.metadata.namespace,URL:.metadata.labels.infrahub-api-url").
		Return([]byte("list result"), nil)

	svc := service.NewCredentialsService(mockCLI)
	err := svc.ListCredentialsSecrets()

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestRemoveCredentialsSecret_Success(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "https://infrahub.example.com").Return("infrahub.example.com", nil)
	mockCLI.On("Hash", "infrahub.example.com").Return("abcd1234")
	mockCLI.On("Delete", mock.Anything, "secret", "infrahub-credentials-abcd1234", "ns").
		Return(nil)

	svc := service.NewCredentialsService(mockCLI)
	err := svc.RemoveCredentialsSecret("https://infrahub.example.com", "ns")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}

func TestRemoveCredentialsSecret_LabelFromURL_Error(t *testing.T) {
	mockCLI := new(mockKubeCLI)
	mockCLI.On("LabelFromURL", "invalid").Return("invalid", errors.New("bad url"))
	mockCLI.On("Hash", "invalid").Return("bad-hash")
	mockCLI.On("Delete", mock.Anything, "secret", "infrahub-credentials-bad-hash", "ns").
		Return(nil)

	svc := service.NewCredentialsService(mockCLI)
	err := svc.RemoveCredentialsSecret("invalid", "ns")

	assert.NoError(t, err)
	mockCLI.AssertExpectations(t)
}
