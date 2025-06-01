package kubecli

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type kubectlCLI struct{}

func NewDefaultKubeCLI() KubeCLI {
	return &kubectlCLI{}
}

func (k *kubectlCLI) ApplyYAML(ctx context.Context, yaml string) error {
	ctx, cancel := setTimeoutIfNoDeadline(ctx, 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (k *kubectlCLI) GetByLabel(ctx context.Context, resource, namespace, labelKey, labelValue string) ([]byte, error) {
	ctx, cancel := setTimeoutIfNoDeadline(ctx, time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"kubectl", "get", resource,
		"-n", namespace,
		"-l", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"-o", "yaml")
	return cmd.Output()
}

func (k *kubectlCLI) ListByLabel(ctx context.Context, resource, label string, outputFormat string) ([]byte, error) {
	ctx, cancel := setTimeoutIfNoDeadline(ctx, time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"kubectl", "get", resource,
		"--all-namespaces",
		"-l", label,
		"-o", outputFormat)
	return cmd.Output()
}

func (k *kubectlCLI) Delete(ctx context.Context, resource, name, namespace string) error {
	ctx, cancel := setTimeoutIfNoDeadline(ctx, time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "delete", resource, name, "-n", namespace)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setTimeoutIfNoDeadline(ctx context.Context, time time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, time)
	}
	return ctx, func() {}
}

func (k *kubectlCLI) EncodeBase64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func (k *kubectlCLI) Hash(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (k *kubectlCLI) LabelFromURL(inputURL string) (string, error) {
	u, err := url.Parse(inputURL)
	if err != nil || u.Scheme == "" {
		return inputURL, err
	}

	hostname := u.Hostname()
	if strings.Contains(hostname, ":") {
		hostname = strings.Split(hostname, ":")[0]
	}

	return hostname, nil
}
