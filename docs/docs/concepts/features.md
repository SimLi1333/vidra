---
title: Features
sidebar_position: 2
---

# Feature Overview

Vidra is a Kubernetes operator designed to manage custom resources with flexibility and safety in complex environments. Below is an overview of its core features and capabilities, including both current functionalities and potential future developments.

---

## Current Features

### Namespace Handling
Vidra tracks and manages resource ownership across namespaces, ensuring consistency and preventing conflicts during the resource lifecycle.

### Safe Handling of Name Changes
Vidra includes logic to safely manage resource renames:
- Detects renamed resources and reconciles to the desired state
- Cleans up old resources to avoid orphaned objects
- Ensures safe reconciliation after name updates

### Multiple `VidraResource` Instances per `InfrahubSync`
Supports creating multiple `VidraResource` instances from a single `InfrahubSync`, enabling:
- Management of multiple resources from one Infrahub artifact
- Simplified configuration without duplicating `InfrahubSync` definitions

### Shared Resource Management
Allows multiple `VidraResource` instances to manage the same Kubernetes resource:
- Enables shared ownership (e.g., namespaces) across different `VidraResource` and `InfrahubSync` configurations
- Prevents deletion of resources still in use by other managers
- Tracks ownership to ensure safe lifecycle operations

### Efficient Caching
Vidra downloads artifacts only if the checksum has changed, reducing unnecessary network calls and improving performance.

### Helm Chart Deployment
Vidra is available as a Helm chart (OCI and standard Helm repository), allowing:
- Installation via `helm repo add` and `helm install`
- Flexible deployment configurations
- Easy integration into Helm-based infrastructure

### Multicluster Support
Vidra supports multi-cluster environments:
- Uses `kubeconfig` contexts to read/write across clusters (stored in Kubernetes Secrets)
- Reconciles resources consistently across multiple environments
- Maintains unique identity and ownership tracking per cluster

### Continuous Delivery of Virtual Machines
Supports continuous delivery of `VirtualMachine` resources via [KubeVirt](https://kubevirt.io):
- Automates creation and updates based on Infrahub artifacts
- Continuously reconciles VM state with the desired configuration

### Event-Driven Reconciliation
Vidra uses an event-driven model to trigger reconciliation:
- Responds immediately to resource changes
- Dynamically adds Informers only for managed resources
- Minimizes overhead while supporting any Kubernetes resource type
- Reduces latency in updates and syncs
- Can be enabled per `InfrahubSync` or globally

### Finalizers for Safe Cleanup
Finalizers ensure that:
- Managed resources are cleaned up if the `VidraResource` is deleted
- Cleanup logic is safely executed

## Experimental Capabilities (Not Yet Fully Validated)

Vidra’s architecture supports advanced scenarios that are available but not yet fully tested:

### Advanced CRs
Potential for managing complex CRs such as:
- Network configurations (e.g., Kubenet, SDC) — Infrahub artifacts could define network policies and configurations
- Cloud-native infrastructure (e.g., Crossplane) — Infrahub artifacts could define Crossplane resources for cloud services
- Any other Kubernetes resource type (CR or not)

These capabilities are technically feasible but require further testing and validation before being considered production-ready.

---

## Future Improvements

### Webhook-Based Reconciliation *(Planned/In Progress)*
Vidra is evolving toward a fully event-based Continuous Integration Operator. Planned webhook support will:
- Trigger real-time reconciliation on Infrahub changes
- Reduce update latency
- Remove reliance on frequent periodic resyncs
- Provide more immediate feedback and state syncing

### Sync to Other Platforms
The `VidraResource` abstraction is independent of Infrahub, allowing for future support of:
- **Helm**: Manage resources via Helm charts
- **Git**: Enable GitOps-style syncing from repositories
- **Other platforms**: Any system providing Kubernetes manifests can be integrated

### Enhanced Observability
Upcoming improvements could include:
- Metrics and logging for operational insights
- Tracing to follow resource lifecycle events

### User Interface
The managed resource and the status of the `InfrahubSync` and `VidraResource` resources could be visualized in a user interface, providing:
- Visualization of resource state and dependencies
- Dashboards for real-time monitoring and management

