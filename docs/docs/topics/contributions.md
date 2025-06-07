---
title: Contributor Guide
sidebar_position: 3
---
import Admonition from '@theme/Admonition';

# Contributor Guide

Contributions to **Vidra Operator** are welcome and appreciated! This guide will help you get started with development and testing.

---

## How It Works

This project follows the [Kubernetes Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses **controllers** that implement a **Reconcile** function. The Reconcile function ensures the actual cluster state moves toward the desired state defined in Kubernetes Custom Resources. The Recooncile function is called with the Custom Resource (CR) as input, which did trigger the Reconcile function.

---

## Development Environment

You’ll need access to a **Kubernetes cluster** for development and testing. The operator uses the current context from your `~/.kube/config` file—whatever cluster `kubectl cluster-info` points to.

<Admonition type="info" title="Local Cluster Recommendation">
We recommend [Kind](https://kind.sigs.k8s.io/) or [Minikube](https://minikube.sigs.k8s.io/) for setting up a local development cluster.
</Admonition>
For help setting up a local cluster, refer to the [Cluster Setup](./deploy) guide.

---

## Prerequisites

Before you begin, install the following tools:

- [Go](https://golang.org/dl/) ≥ 1.24
- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Operator SDK](https://sdk.operatorframework.io/docs/installation/)

---

## Getting Started

```shell
# 1. Clone the repository
git clone https://github.com/infrahub-operator/vidra.git
cd vidra

# 2. Set your kubeconfig context to point to your dev cluster
kubectl config use-context <your-context>

# 3. Install the CRDs
make install

# 4. Run the operator locally
make run
````

<Admonition type="note" title="Running Locally"> Running locally uses your machine’s Go environment and the active kubeconfig context. It does not apply in-cluster RBAC permissions, so make sure your user has sufficient access. </Admonition>

---

## Testing Changes
To test your changes, you can use the following commands:

```shell
# 1. Run tests
make test

# 2. Run integration tests
make test-e2e

# 3. Run the operator with your changes
make run

# 4. Lint the code
make lint
```
---

## Running Vidra Operator on the Cluster

To test the operator inside the cluster (with RBAC and in a real-world environment):
```shell
# 1. Build the operator image
make docker-build IMG=<some-registry>/vidra:<tag>

# 2. Push the image to your container registry
make docker-push IMG=<some-registry>/vidra:<tag>

# 3. Deploy the operator
make deploy IMG=<some-registry>/vidra:<tag>

# 4. Check the operator logs    
kubectl logs -l app.kubernetes.io/name=vidra-operator -n vidra-system
```

<Admonition type="hint" title="Hint">
Run `make --help` to see all available automation commands
```shell
make --help
```
</Admonition>
---

Further Reading

[Operator SDK Documentation](https://sdk.operatorframework.io/docs/)
[Kubebuilder Book](https://book.kubebuilder.io/)
[Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)

