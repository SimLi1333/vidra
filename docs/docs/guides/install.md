---
title: Installing Vidra
sidebar_position: 1
---
# User Guide
## Install the Vidra Operator using Helm:
### Installation Pre-requisites

| Tool                        | Version   | Installation | Description                                                                                                   |
|-----------------------------|-----------|--------------|---------------------------------------------------------------------------------------------------------------|
| Kubernetes                  | ^1.26.0   | [Installation](https://kubernetes.io/docs/setup/) | Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications. |
| Helm                        | ^3.0.0    | [Installation](https://helm.sh/docs/intro/install/) | Helm is a package manager for Kubernetes that helps you manage Kubernetes applications. It uses a packaging format called charts. |

### Helm Installation
To install the Vidra Operator using Helm, first add the Infrahub Operator Helm repository and then install the Vidra chart:

```sh
helm repo add infrahub-operator https://infrahub-operator.github.io/vidra
helm repo update
helm install vidra infrahub-operator/vidra --namespace vidra-system --create-namespace
```

---

## Install the Vidra Operator using OLM:
### Installation Pre-requisites

| Tool                        | Version   | Installation | Description                                                                                                   |
|-----------------------------|-----------|--------------|---------------------------------------------------------------------------------------------------------------|
| Kubernetes                  | ^1.26.0   | [Installation](https://kubernetes.io/docs/setup/) | Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications. |
| Operator Lifecycle Manager  | v0.32.0   | [Installation](https://operator-framework.github.io/operator-lifecycle-manager/docs/installation/) | OLM is a Kubernetes project that helps you manage the lifecycle of operators running on your cluster. |

**Alternative OLM installation:**

```sh
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.32.0/install.sh | bash -s v0.32.0
```

## Operator Lifecycle Manager (OLM) Installation

Install the vidra Operator by creating a catalog source and subscription:

```sh
kubectl apply -f https://raw.githubusercontent.com/Infrahub-Operator/Vidra/main/install/catalogsource.yaml -f https://raw.githubusercontent.com/Infrahub-Operator/Vidra/main/install/subscription.yaml
```

Wait for the Vidra Operator to be installed (this might take a few seconds):

```sh
kubectl get csv -n operators -w
```
