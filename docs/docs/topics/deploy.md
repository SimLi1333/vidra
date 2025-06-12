---
title: Cluster Setup
sidebar_position: 5
---
import Admonition from '@theme/Admonition';

# Development Cluster Setup

To develop and test the **Vidra Operator** locally, create a Kubernetes cluster using either **minikube** or **kind**. Both tools provide a simple way to run Kubernetes on your machine for day-to-day development tasks.

| Local Solution | Quick-Start Guide |
| -------------- | ---------------- |
| **minikube**   | https://minikube.sigs.k8s.io/docs/start/ |
| **kind**       | https://kind.sigs.k8s.io/docs/user/quick-start/ |

---
## Infrahub

### Installation

To install Infrahub in your local cluster, follow the [Infrahub installation guide](https://docs.infrahub.app/guides/installation).

<Admonition type="note" title="Infrahub Installation">
We recommend [Docker Compose](https://docs.infrahub.app/guides/installation#docker-compose) for local Infrahub installations. It simplifies the setup process and is well-suited for development environments.
For production deployments, consider using the [Helm chart repo](https://github.com/opsmill/infrahub-helm).
</Admonition>

### Configuration

To configure Infrahub for Vidra, ensure Infrahub generates the necessary artifacts with valid Kubernetes manifests. 
You can get an idea of how to configure Infrahub for Vidra by looking at our guide: [Preparing Infrahub for Vidra](../guides/infrahub).

### KubeVirt

If you want your cluster to be able to run virtual machines, you can install [KubeVirt](https://kubevirt.io/). This is not required for Vidra, but it is a nice addition to be able to create and manage virtual machines from Infrahub.

<Admonition type="note" title="KubeVirt Installation">
We tested Infrahub and Vidra with [KubeVirt on a Talos Linux cluster](https://www.talos.dev/v1.10/advanced/install-kubevirt/).  
While other Kubernetes distributions probably work, we recommend Talos.
</Admonition>

---
