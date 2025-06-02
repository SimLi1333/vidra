package service

import (
	"context"
	"fmt"
	"os"
	"vidra-cli/internal/adapter/kubecli"
)

type infrahubSyncService struct {
	kubecli kubecli.KubeCLI
}

func NewInfrahubSyncService(cli kubecli.KubeCLI) InfrahubSyncService {
	return &infrahubSyncService{kubecli: cli}
}

func (s *infrahubSyncService) PrintInfrahubSync(url, namespace, name string) error {
	host, err := s.kubecli.LabelFromURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
		fmt.Fprintf(os.Stderr, "trying input %s as URL\n", host)
	}
	yaml, err := s.kubecli.GetByName(context.Background(), "infrahubSync", namespace, name)
	if err != nil {
		return err
	}
	fmt.Println(string(yaml) + "\n\n---\n")
	return nil
}

func (s *infrahubSyncService) ApplyInfrahubSync(url, artifact, branch, date, server, destNamespace string, reconcileOnEvent bool, namespace, name string) error {
	host, err := s.kubecli.LabelFromURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
		fmt.Fprintf(os.Stderr, "trying input %s as URL\n", host)
	}

	yaml := generateInfrahubSyncYAML(url, artifact, branch, date, server, destNamespace, reconcileOnEvent, namespace, name)
	fmt.Print(yaml + "\n\n---\n")
	return s.kubecli.ApplyYAML(context.Background(), yaml)
}

func (s *infrahubSyncService) ListInfrahubSync() error {
	result, err := s.kubecli.ListByLabel(
		context.Background(),
		"infrahubsyncs",
		"",
		"custom-columns=InfrahubSync-NAME:.metadata.name,NAMESPACE:.metadata.namespace,URL:.spec.source.infrahubAPIURL,ARTIFACT:.spec.source.artefactName,BRANCH:.spec.source.targetBranch,DATE:.spec.source.targetDate,DESTINATION_SERVER:.spec.destination.server,DESTINATION_NAMESPACE:.spec.destination.namespace,RECONCILE_ON_EVENTS:.spec.destination.reconcileOnEvents",
	)
	fmt.Println(string(result) + "\n\n---\n")
	return err
}

func (s *infrahubSyncService) RemoveInfrahubSync(url, namespace, name string) error {
	host, err := s.kubecli.LabelFromURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse URL: %v\n", err)
		fmt.Fprintf(os.Stderr, "trying input %s as URL\n", host)
	}
	return s.kubecli.Delete(context.Background(), "infrahubSync", name, namespace)
}

func generateInfrahubSyncYAML(url, artifact, branch, date, server, destNamespace string, reconcileOnEvent bool, namespace, name string) string {
	yaml := fmt.Sprintf(`apiVersion: infrahub.operators.com/v1alpha1
kind: InfrahubSync
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/name: vidra
    app.kubernetes.io/managed-by: kustomize
spec:
  source:
`, name, namespace)

	if url != "" {
		yaml += fmt.Sprintf("    infrahubAPIURL: %s\n", url)
	}
	if branch != "" {
		yaml += fmt.Sprintf("    targetBranch: %s\n", branch)
	}
	if date != "" {
		yaml += fmt.Sprintf("    targetDate: %s\n", date)
	}
	if artifact != "" {
		yaml += fmt.Sprintf("    artefactName: %s\n", artifact)
	}

	yaml += "  destination:\n"
	if server != "" {
		yaml += fmt.Sprintf("    server: %s\n", server)
	}
	if destNamespace != "" {
		yaml += fmt.Sprintf("    namespace: %s\n", destNamespace)
	}
	yaml += fmt.Sprintf("    reconcileOnEvents: %t\n", reconcileOnEvent)

	return yaml
}
