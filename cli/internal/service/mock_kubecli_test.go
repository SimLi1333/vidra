package service_test

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockKubeCLI struct {
	mock.Mock
}

func (m *mockKubeCLI) GetByLabel(ctx context.Context, kind, namespace, labelKey, labelValue string) ([]byte, error) {
	args := m.Called(ctx, kind, namespace, labelKey, labelValue)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockKubeCLI) EncodeBase64(data string) string {
	args := m.Called(data)
	return args.String(0)
}

func (m *mockKubeCLI) Hash(data string) string {
	args := m.Called(data)
	return args.String(0)
}

func (m *mockKubeCLI) LabelFromURL(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *mockKubeCLI) ApplyYAML(ctx context.Context, yaml string) error {
	args := m.Called(ctx, yaml)
	return args.Error(0)
}

func (m *mockKubeCLI) ListByLabel(ctx context.Context, kind, label, format string) ([]byte, error) {
	args := m.Called(ctx, kind, label, format)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockKubeCLI) Delete(ctx context.Context, kind, name, namespace string) error {
	args := m.Called(ctx, kind, name, namespace)
	return args.Error(0)
}
func (m *mockKubeCLI) GetByName(ctx context.Context, kind, namespace, name string) ([]byte, error) {
	args := m.Called(ctx, kind, namespace, name)
	return args.Get(0).([]byte), args.Error(1)
}
