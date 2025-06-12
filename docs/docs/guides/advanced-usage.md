---
title: Advanced Usage
sidebar_position: 4
---
import Admonition from '@theme/Admonition';

<Admonition type="note" title="Note">
Most of these configurations can also be done using the CLI tool for Vidra. The CLI tool provides a convenient way to manage Vidra resources and configurations. You can find more information about the CLI tool in the [CLI documentation](/cli).
</Admonition>

# Advanced Usage of Vidra Operator
This guide covers advanced usage scenarios for the Vidra Operator, including multi-cluster synchronization and `VidraResource` management.

If you want to synchronize resources to a different Kubernetes cluster, you can specify the `server` field in the `destination` section of the `InfrahubSync` resource. This allows Vidra to connect to the specified cluster and apply the resources there.

For this to work, you need to create a Kubernetes Secret containing the kubeconfig for the target cluster. The Secret should have the label `cluster-kubeconfig` with the URL of the target cluster. Below is an example of how to create this Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-kubeconfig
  labels:
    cluster-kubeconfig: <cluster-url-without-http>
type: Opaque
data:
  kubeconfig: <base64-encoded-kubeconfig>
```
<Admonition type="note" title="Note">
The `kubeconfig` field should contain the kubeconfig for the target cluster, encoded in base64. You can use the following command to encode your kubeconfig:

**MacOS/Linux:**
```sh
echo -n <kubeConfig> | base64 >> temp.txt
```
or 
```sh
base64 -w 0 -i <path to kubeConfig>
```

**Windows:**
```Python
python -c 'import base64, os; print(base64.b64encode(open(os.path.expanduser("<path to kubeConfig.conf>"), "rb").read()).decode())'
```
</Admonition>

---

## Creating a `VidraResource`
Usually you will use the [`InfrahubSync`](usage##creating-an-infrahubsync-resource) resource to synchronize resources from Infrahub to Kubernetes. However, if you want to create your own workflow or test just the kubernetes part of Vidra, you can use the `VidraResource` CR. This allows you to define a manifest in the `VidraResource` and Vidra will do its job by reconciling the desired resources to the specified destination cluster and monitor the `VidraResource` and all its managed resourges for change, keeping the `VidraResource.spec.manifest` in sync with the managed resources.

To manage resources using the Vidra Operator, you can create a `VidraResource`. Below is an example of how a `VidraResource` looks like:

```yaml
apiVersion: infrahub.operators.com/v1alpha1
kind: VidraResource
metadata:
  labels:
    app.kubernetes.io/name: vidra
    app.kubernetes.io/managed-by: kustomize
  name: vidraresource-sample
spec:
    destination:
      server: https://kubernetes.default.svc
      namespace: default
      reconcileOnEvents: true
    manifest: '{"apiVersion": "v1","kind": "Namespace","metadata":{"name": "ns-sample"}}'
```

A `VidraResource` only requires the `spec.manifest` field, which should contain the Kubernetes manifests you want Vidra to manage.

A `VidraResource` created by `InfrahubSync` and a completed reconciliation will look like this:
```yaml
apiVersion: infrahub.operators.com/v1alpha1
kind: VidraResource
metadata:
  creationTimestamp: "2025-06-01T00:59:48Z"
  finalizers:
  - vidraresource.infrahub.operators.com/finalizer
  generation: 2
  name: 183c2696-244a-db4d-3835-c51c357bbaf3
  ownerReferences:
  - apiVersion: infrahub.operators.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: InfrahubSync
    name: sync-test-multi-x
    uid: eb9fe47d-73cb-4a5a-9b2c-49d6bf89a7f9
  resourceVersion: "89703"
  uid: 777c45c3-382f-4346-876a-185706d06fb4
spec:
  destination:
    namespace: default
    reconcileOnEvents: true
  manifest: |-
    ---
    apiVersion: v1
    kind: Namespace
    metadata:
      name: ns-example

    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: dep-example
      namespace: ns-example
      labels:
        app: l-example
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: l-example
      template:
        metadata:
          labels:
            app: l-example
        spec:
          containers:
            - name: con-example
              image: public.ecr.aws/pahudnet/nyancat-docker-image
              ports:
                - containerPort: 80

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: svc-example
      namespace: ns-example
      labels:
        app: l-example
    spec:
      ports:
        - port: 80
          name: http
      selector:
        app: l-example

    ---
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: ing-example
      namespace: ns-example
    spec:
      rules:
        - host: demo.cldop-test-0.network.garden
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: svc-example
                    port:
                      number: 80
  reconciledAt: "2025-06-01T00:59:48Z"
status:
  DeployState: Succeeded
  lastSyncTime: "2025-06-01T00:59:48Z"
  managedResources:
  - apiVersion: v1
    kind: Namespace
    name: ns-example
  - apiVersion: apps/v1
    kind: Deployment
    name: dep-example
    namespace: ns-example
  - apiVersion: v1
    kind: Service
    name: svc-example
    namespace: ns-example
  - apiVersion: networking.k8s.io/v1
    kind: Ingress
    name: ing-example
    namespace: ns-example
```

The managed resources will have the following labels and annotations set by Vidra:
```yaml
  annotations:
    managed-by: vidra
    vidraresource.infrahub.operators.com/owned-by: 183c2696-244a-db4d-3835-c51c357bbaf3
  labels:
    kubernetes.io/metadata.name: ns-example
    managed-by: vida
  ownerReferences:
  - apiVersion: infrahub.operators.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: VidraResource
    name: 183c2696-244a-db4d-3835-c51c357bbaf3
    uid: 777c45c3-382f-4346-876a-185706d06fb4
---