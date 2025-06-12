---
title: Comparison
sidebar_position: 1
---
import Admonition from '@theme/Admonition';

# Comparison to ArgoCD and Flux

Vidra is one of several open-source tools designed to manage and deploy Kubernetes resources. This comparison provides a concise overview of the key features and philosophies behind ArgoCD and Flux, highlighting where Vidra fits in and what unique value it offers to users.

## ArgoCD

**ArgoCD** is a declarative, GitOps continuous delivery tool for Kubernetes. It watches a Git repository and ensures the desired application state described in Git is reflected in the cluster.

- **GitOps-first**: ArgoCD embraces the GitOps model, making Git the single source of truth for application state.
- **UI and CLI**: Comes with a rich user interface and CLI for observing and managing deployments.
- **Application-centric model**: ArgoCD uses an "Application" CR to represent and manage deployments.
- **Multi-tenancy**: Built-in support for managing multiple teams and access controls.
- **Custom resource sync**: Can sync not only Helm charts and Kustomize, but also plain manifests and custom plugins.

**Limitations compared to Vidra**:
- Hard requirement to sync with a Git repository, which does not suit our use case.
- GitOps workflow is tightly coupled to Git — limited flexibility for non-Git workflows or programmatic triggers.
- Integrating with systems outside of Git (e.g., artifact registries, APIs) often requires additional tooling.

## Flux

**Flux** is another GitOps operator for Kubernetes that focuses on simplicity, modularity, and composability.

- **Toolkit-based**: Flux is a set of composable controllers (e.g., `source-controller`, `kustomize-controller`), like Vidra aims to be.
- **High modularity**: Encourages building your own GitOps pipeline with only the pieces you need.
- **Secure by design**: Follows Kubernetes RBAC closely and supports scoped access.
- **Built for automation**: Well-suited for machine-driven workflows and automated environments.

**Limitations compared to Vidra**:
- Less visibility into application-level deployments unless additional tooling (e.g., Weave GitOps UI) is used.
- Limited out-of-the-box support for non-Git sources unless extended.
- No multicluster support out of the box.

## Vidra

**Vidra** is designed as a lightweight deployment orchestrator for Kubernetes that emphasizes pluggability, programmatic integration, and simplicity.

- **No knowledge about CR required**: Vidra can be managed over the Vidra CLI; no knowledge required about the CR.
- **Pluggable execution engine**: Custom workflows and deployment logic can be added without rebuilding the core system.
- **Lightweight footprint**: Minimal operational overhead, suitable for embedding into broader automation frameworks.

**Limitations**:
- Only supports Infrahub as of today.

**Unique strengths**:
- Better suited for scenarios where Git is not the central source of truth.
- Easier to integrate with existing systems and APIs.

---

Vidra is not a direct replacement for ArgoCD or Flux, especially as it is the only one connecting to Infrahub. Instead, it complements them by providing a lightweight, pluggable solution for users who need more flexibility in managing Kubernetes resources without being tied to a Git-centric workflow.
