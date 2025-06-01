package service

import (
	"context"
	"fmt"
	"vidra-cli/internal/adapter/kubecli"
)

type configService struct {
	kubecli kubecli.KubeCLI
}

func NewConfigService(cli kubecli.KubeCLI) ConfigService {
	return &configService{kubecli: cli}
}

func (s *configService) PrintConfigMap(namespace string) error {
	yaml, err := s.kubecli.GetByLabel(context.Background(), "configmap", namespace, "app", "vidra")
	if err != nil {
		return err
	}
	fmt.Println(string(yaml) + "\n\n---\n")
	return nil
}

func (s *configService) ApplyConfigMap(requeueSyncAfter, requeueResourceAfter, queryName, namespace string) error {
	yaml := generateConfigMap(namespace, requeueSyncAfter, requeueResourceAfter, queryName)
	fmt.Println(yaml + "\n\n---\n")
	return s.kubecli.ApplyYAML(context.Background(), yaml)
}

func (s *configService) ListConfigMaps() error {
	result, err := s.kubecli.ListByLabel(
		context.Background(),
		"configmaps",
		"app=vidra",
		"custom-columns=CONFIGMAP-NAME:.metadata.name,NAMESPACE:.metadata.namespace,REQUEUE_SYNC_AFTER:.data.requeueSyncAfter,REQUEUE_RESOURCE_AFTER:.data.requeueResourceAfter,QUERY_NAME:.data.queryName",
	)
	if err != nil {
		return err
	}
	fmt.Println(string(result) + "\n\n---\n")
	return nil
}

func (s *configService) RemoveConfigMap(namespace string) error {
	return s.kubecli.Delete(context.Background(), "configmap", "vidra-config", namespace)
}

func generateConfigMap(ns, syncDuration, resourceDuration, query string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: vidra-config
  namespace: %s
  labels:
    app: vidra
data:
  requeueSyncAfter: "%s"
  requeueResourceAfter: "%s"
  queryName: "%s"
`, ns, syncDuration, resourceDuration, query)
}
