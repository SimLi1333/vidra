---
title: Using Vidra
sidebar_position: 3
---
import Admonition from '@theme/Admonition';

## Introduction

This guide will walk you through the basic configuration and usage of Vidra.

<Admonition type="warning" title="Warning">
A running Kubernetes cluster with the Vidra operator installed is required. If you haven't installed the Vidra operator yet, please refer to the [installation guide](/guides/install.md).
</Admonition>
<Admonition type="note" title="Note">
All these configurations can also be done using the CLI tool for Vidra. The CLI tool provides a convenient way to manage Vidra resources and configurations. You can find more information about the CLI tool in the [CLI documentation](/cli).
</Admonition>

## Configuring Vidra

Vidra uses a ConfigMap to manage its configuration. Below is an example ConfigMap you can use as a starting point:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vidra-config
  labels:
    app: vidra # Important: This label is used to identify the operator's resources.
  namespace: vidra-system
data:
  requeueSyncAfter: "1m" # How often Vidra syncs with Infrahub. (if you do not want to use the default value of 1 minute)
  requeueResourcesAfter: "1m" # How often managed resources are reconciled. (if you do not want to use the default value of 10 minutes)
  queryName: "ArtifactIDs" # Infrahub GraphQL query name for getting Artifact IDs. (if you do not want to use the default value of "ArtifactIDs")
  eventBasedReconcile: "true" # Enable event-based reconciliation. (default is false)
```
<Admonition type="note" title="Note">
All the fields in the ConfigMap are optional and can be customized according to your needs. If you do not specify a field, Vidra will use its default values.
</Admonition>

`Requeue` values are specified as positive duration strings. A duration is a sequence of decimal numbers with optional fractions and a unit suffix, such as "300ms" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", and "h". Negative durations are not allowed and will be ignored. Setting a value to `0` disables requeueing for that resource (e.g., for maintenance).

`requeueResourcesAfter` is disabled if you set `eventBasedReconcile: "true"`, as it will use the Kubernetes event system to trigger reconciliations instead of a time-based requeue.

<Admonition type="note" title="Note">
If you want to use a different namespace than `vidra-system`, make sure to adjust the `namespace` field in the metadata section accordingly. Vidra will find the ConfigMap based on the label `app: vidra`.
</Admonition>

### Applying the ConfigMap

To apply the configuration, save the above YAML to a file (e.g., `vidra-config.yaml`) and run:

```sh
kubectl apply -f vidra-config.yaml
```

---

## Creating an `infrahub-credentials` Secret

To authenticate with your Infrahub, you need to create a Kubernetes Secret containing your Infrahub API credentials. The Secret should contain the label `infrahub-api-url` with the URL of your Infrahub instance. Below is an example of how to create this Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: infrahub-credentials
  namespace: vidra-system
  labels:
    infrahub-api-url: "198.19.248.5" # Important: This label is used to identify the operator's resources.
data:
  username: YWRtaW4=
  password: aW5mcmFodWI=
```

<Admonition type="note" title="Note">
The `username` and `password` fields should contain your Infrahub API credentials, encoded in base64. You can use the following command to encode your credentials:

**macOS/Linux:**
```sh
echo -n <value> | base64
```
**Windows:**
```Python
python -c 'import base64; print(base64.b64encode(b"<value>").decode())'
```
</Admonition>

The label `infrahub-api-url` is used to identify the Infrahub instance that the Vidra operator should connect to. Make sure to replace the value with your actual Infrahub API URL or IP.  
**Example:** The URL must look like `infrahub-server.infrahub.orb.local` if your Infrahub is accessible at `https://infrahub-server.infrahub.orb.local`, since Kubernetes does not allow `/` in labels.

<Admonition type="note" title="Note">
If you need to connect to different Infrahub instances, you can create multiple Secrets with different `infrahub-api-url` labels. The Vidra operator will use the Secret that matches the label of the `InfrahubSync` resource.
</Admonition>

<Admonition type="note" title="Note">
If you want to use a different namespace than `vidra-system`, make sure to adjust the `namespace` field in the metadata section accordingly. Vidra will find the Secret based on the label `infrahub-api-url`.
</Admonition>

### Applying the Secret

To apply the configuration, save the above YAML to a file (e.g., `infrahub-secret.yaml`) and run:

```sh
kubectl apply -f infrahub-secret.yaml
```
---

## Creating an `InfrahubSync` Resource

To synchronize resources from Infrahub, you need to create an `InfrahubSync` resource. Below is an example of how to create this resource:

```yaml
apiVersion: infrahub.operators.com/v1alpha1
kind: InfrahubSync
metadata:
  name: sync-test-webserver
  labels:
    app.kubernetes.io/name: vidra
    app.kubernetes.io/managed-by: kustomize
spec:
  source:
    # The URL of your Infrahub instance.
    infrahubAPIURL: "https://infrahub-server.infrahub.orb.local"
    # The branch in Infrahub to query for Artifacts. (Optional)
    targetBranch: "main"
    # The date to query for Artifacts. If not set, the latest branch is used. (Optional)
    targetDate: "2025-04-09T00:00:00Z"
    # Name of the Artifact Definition in Infrahub to query for Artifacts containing k8s manifests.
    artefactName: "Webserver_Manifest"
  destination:
    # The URL of the Kubernetes cluster where the resources should be applied (Multi-cluster mode). If set to "https://kubernetes.default.svc" or not set at all, the current cluster is used. (Optional)
    server: 'https://k8s-cldop-test-0.network.garden:6443'
    # The namespace in the destination cluster which is used as fallback if the managed resources do not have a namespace defined. If not set, the default namespace is used. (Optional)
    namespace: 'default'
    # If set to true, all managed resources in this sync will be reconciled on events (e.g., creation, update, deletion) instead of a time-based requeue. Default is false. (Optional)
    reconcileOnEvents: true
```
<Admonition type="note" title="Note">
If you want to synchronize multiple Artifact Definitions (like Webserver and VirtualMachines), you can create multiple `InfrahubSync` resources with different `artefactName` values.
</Admonition>

<Admonition type="note" title="Note">
If you want to use a different namespace, make sure to add the `namespace` field in the metadata section accordingly. Vidra will find the `InfrahubSync` resource based on its Kind.
</Admonition>

### Applying the InfrahubSync Resource

To apply the `InfrahubSync` resource, save the above YAML to a file (e.g., `infrahub-sync.yaml`) and run:

```sh
kubectl apply -f infrahub-sync.yaml
```

The `InfrahubSync` resource will trigger the Vidra operator to start synchronizing resources from Infrahub based on the specified query and parameters. You can monitor the status of the synchronization by checking the status of the `InfrahubSync` resource:

```sh
kubectl get infrahubsync sync-test-webserver -o jsonpath='{.status}'
```

<Admonition type="note" title="Note">
If you use the `destination.server` field to specify a different Kubernetes cluster, make sure to create a Kubernetes Secret with the kubeconfig for that cluster, as described in the [Multi-Cluster Mode](advanced-usage#multi-cluster-mode) section.
</Admonition>

<Admonition type="note" title="Note">
Some sample resources for Vidra can be found in the [config samples directory](https://github.com/infrahub-operator/vidra/tree/main/config/samples).
</Admonition>

---
