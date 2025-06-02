package kubecli

import "context"

type KubeCLI interface {
	ApplyYAML(ctx context.Context, yaml string) error
	GetByLabel(ctx context.Context, resource, namespace, labelKey, labelValue string) ([]byte, error)
	GetByName(ctx context.Context, resource, namespace, name string) ([]byte, error)
	ListByLabel(ctx context.Context, resource, label string, outputFormat string) ([]byte, error)
	Delete(ctx context.Context, resource, name, namespace string) error
	EncodeBase64(data string) string
	Hash(data string) string
	LabelFromURL(url string) (string, error)
}
