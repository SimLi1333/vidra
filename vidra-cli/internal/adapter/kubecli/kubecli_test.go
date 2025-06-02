package kubecli_test

import (
	"testing"

	"vidra-cli/internal/adapter/kubecli"
)

func TestEncodeBase64(t *testing.T) {
	cli := kubecli.NewDefaultKubeCLI()

	input := "my-secret"
	expected := "bXktc2VjcmV0"

	result := cli.EncodeBase64(input)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestHash(t *testing.T) {
	cli := kubecli.NewDefaultKubeCLI()

	input := "example.com"
	expected := "5ababd603b22780302dd8d83498e5172"
	result := cli.Hash(input)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestLabelFromURL(t *testing.T) {
	cli := kubecli.NewDefaultKubeCLI()

	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com", "example.com"},
		{"https://example.com:8080", "example.com"},
		{"example.com", "example.com"},
		{"http://user:pass@example.com", "example.com"},
	}

	for _, tt := range tests {
		host, err := cli.LabelFromURL(tt.url)
		if host != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, host)
		}
		if tt.url == "example.com" && err != nil {
			t.Errorf("expected no error for plain URL input, got: %v", err)
		}
	}
}
