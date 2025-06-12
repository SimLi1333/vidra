---
title: Concepts
sidebar_position: 1
---

## Infrahub

**Infrahub** is a inventory platform that stores information about infrastructure components such as servers, services, environments, or configurations. It acts as a centralized source of truth for discovering and querying infrastructure relationships and attributes, enabling use cases like documentation, automation, and auditing. We mainly use vidra to automate the creation of kubernetes manifests based on the data stored in Infrahub.

## Vidra Operator

The **Vidra Operator** is a Kubernetes Operator responsible for keeping Kubernetes resources in sync with inventory data and its manifests stored in Infrahub. It automates the creation and update of Kubernetes native resources and Custom Resources (CRs) based on artifacts retrieved from Infrahub queries. Vidra enables continious delivery into the Kubernetes control plane.

## Artifact

An **Artifact** in this context refers to a structured result (usually JSON or YAML) retrieved from Infrahub using a query. These artifacts may describe infrastructure entities such as virtual machines, web-services, or any other kubernetes resources, and are the source data Vidra uses to drive continious delivery of Kubernetes resources.

## InfrahubSync

An `InfrahubSync` is a Kubernetes Custom Resource that defines a desired sync operation between Infrahub and Kubernetes. It contains:

- A **source** section including:
    - The **Infrahub API URL** to connect to.
    - The **target branche** and **Target Date** to specify which version of the Artifact to use.
    - A **artifact name** to select the corect data from infrahub.
- A **destination** section specifying:
    - The **Kubernetes Server URL** (optional, defaults to the current cluster).
    - The **Kubernetes namespace** where the resources should be created.
    - A flag to enable **reconciliation on events** (optional, defaults to false).
- A **status** section showing the current state of the sync operation.

The Vidra Operator watches these resources and ensures the corresponding `VidraResource`s are created or updated as needed.

## VidraResource

An `VidraResource` is a Kubernetes CR created and managed by the Vidra Operator based on the results of an `InfrahubSync`. Each `VidraResource` corresponds to a specific Artifact from Infrahub and contains:


- The **manifest** field containing the content of the artifact, representing the actual resource definition (usually in JSON or YAML format).
- The **destination** section like in `InfrahubSync`, specifying where the resource should be applied.
- A **status** section showing the reconciliation status.

These CRs act as structured mirrors of the infrastructure state described in Infrahub.

## Reconciliation

Vidra’s reconciliation process runs when:

- An `InfrahubSync` is created or updated.
- The scheduled resync interval is triggered.
- An `VidraResource` is created or updated.
- An event occurs on a managed resource (if reconciliation on events is enabled). This is similar to autoHeal.

During reconciliation, Vidra authenticates with Infrahub, fetches the specified artifact, parses it, and ensures Kubernetes resources match the current state on the destination cluster.

## Managed Resources
Managed resources are the Kubernetes resources created or updated by Vidra based on the `VidraResource` manifests, downloaded from infrahub. These resources can include any Kubernetes object type, such as:
- Deployments
- Services
- Namespaces
- ConfigMaps
- Custom Resources like `VirtualMachine` etc.

## Code Structure

Vidra separates concerns through a **clean architecture** structure, where domain logic is kept independent of Kubernetes and Infrahub-specific implementation details. 

The domain model includes types like:

- `Artifact`

The `Manifest` itselfe is not saved in the domain model, but rather in the `VidraResource` CR. The domain model focuses on the core logic and operations related to artifacts. 

This promotes testability and modularity.

## Links

To better understand the underlying technologies used in Vidra, refer to the following resources:

- [Kubernetes](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/)
- [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Infrahub](https://docs.infrahub.dev/)  
- [Operator SDK](https://sdk.operatorframework.io/docs/)
