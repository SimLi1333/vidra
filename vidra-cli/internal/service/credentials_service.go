package service

import (
	"context"
	"fmt"
	"os"
	"vidra-cli/internal/adapter/kubecli"
)

type credentialsService struct {
	kubecli kubecli.KubeCLI
}

func NewCredentialsService(cli kubecli.KubeCLI) CredentialsService {
	return &credentialsService{kubecli: cli}
}

func (s *credentialsService) PrintCredentialsSecret(url, namespace string) error {
	host, err := s.kubecli.LabelFromURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
		fmt.Fprintf(os.Stderr, "trying input %s as URL\n", host)
	}
	yaml, err := s.kubecli.GetByLabel(context.Background(), "secret", namespace, "infrahub-api-url", host)
	if err != nil {
		return err
	}
	fmt.Println(string(yaml) + "\n\n---\n")
	return nil
}

func (s *credentialsService) ApplyCredentialsSecret(url, username, password, namespace string) error {
	hostname, err := s.kubecli.LabelFromURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
		fmt.Fprintf(os.Stderr, "trying input %s as URL\n", hostname)
	}
	credentialsID := s.kubecli.Hash(hostname)
	yaml := generateCredentialsSecretYAML(credentialsID, namespace, hostname, s.kubecli.EncodeBase64(url), s.kubecli.EncodeBase64(username), s.kubecli.EncodeBase64(password))
	fmt.Print(yaml + "\n\n---\n")
	return s.kubecli.ApplyYAML(context.Background(), yaml)
}

func (s *credentialsService) ListCredentialsSecrets() error {
	result, err := s.kubecli.ListByLabel(
		context.Background(),
		"secrets",
		"infrahub-api-url",
		"custom-columns=SECRET-NAME:.metadata.name,NAMESPACE:.metadata.namespace,URL:.metadata.labels.infrahub-api-url",
	)
	fmt.Println(string(result) + "\n\n---\n")
	return err
}

func (s *credentialsService) RemoveCredentialsSecret(url, namespace string) error {
	host, err := s.kubecli.LabelFromURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
		fmt.Fprintf(os.Stderr, "trying input %s as URL\n", host)
	}
	hash := s.kubecli.Hash(host)
	return s.kubecli.Delete(context.Background(), "secret", fmt.Sprintf("infrahub-credentials-%s", hash), namespace)
}

func generateCredentialsSecretYAML(credentialsID, ns, trimmedUrl, url, username, password string) string {
	// Generate the Secret YAML
	return fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: infrahub-credentials-%s
  namespace: %s
  labels:
    infrahub-api-url: "%s"
data:
  username: %s
  password: %s
  infrahubapiurl: %s`,
		credentialsID,
		ns,
		trimmedUrl,
		username,
		password,
		url,
	)
}
