---
title: Architecture
sidebar_position: 1
---

## Introduction

This chapter presents the architectural design of **Vidra**, a Kubernetes operator developed to automate the synchronization and management of infrastructure resources via the Infrahub platform. Vidra streamlines the process of querying, retrieving, and applying configuration artifacts to a Kubernetes cluster, enabling efficient, automated infrastructure lifecycle management.

The architecture is designed to be modular, extensible, and resilient, leveraging Kubernetes operator patterns and cloud-native best practices. This chapter details the core components, data flows, interactions, and design decisions underlying Vidra.

---

## 1. Overview of Vidra Architecture

Vidra is implemented as a **Kubernetes operator**, extending the Kubernetes control plane to manage custom resources representing Infrahub synchronization configurations. The operator continuously reconciles desired states defined in these custom resources with actual cluster state, ensuring automated, declarative management of infrastructure artifacts.

### Key Architectural Goals:

- **Declarative Resource Management:** Users specify synchronization requirements via custom resources; Vidra automates downstream artifact application.
- **Extensibility:** Modular architecture enables support for additional artifact types, expanded Infrahub query capabilities, and even integration with any other system that generate manifests.
- **Scalability and Robustness:** Efficient handling of concurrent sync operations with robust error handling.
- **Separation of Concerns:** Clear boundaries between domain logic, external system integration, and Kubernetes API interactions.

---

## 2. High-Level Components

Vidra consists of two primary layers, implemented as Go packages:

### 2.1 Controller Layer

- **Reconciler:** The heart of Vidra, this component watches for changes to `InfrahubSync` custom resources. Upon detection, it triggers synchronization workflows to update cluster state.
   - Encapsulates the core synchronization logic: authenticating with Infrahub, running queries, downloading artifacts, and interpreting their contents.
   - Contains orchestration for multi-step processes, retries, and state transitions.
   - Implements interfaces to abstract external dependencies.

### 2.3 Adapter Layer

- **Infrahub Adapter:** Handles communication with the Infrahub API — login, query execution, artifact download.
- **Kubernetes Adapter:** Reads and writes Kubernetes resources, manages manifests, and updates cluster state.
- **Artifact Processing:** Decodes artifact content (e.g., YAML manifests) and prepares them for application.

---

## 3. Data Flow and Interaction

1. **Custom Resource Definition (CRD):** Users create an `InfrahubSync` resource specifying parameters such as the Infrahub query, target artifact, target Git branch, and sync schedule.

2. **Event Trigger:** The Kubernetes API server notifies Vidra’s controller of the resource creation or update.

3. **Reconciliation Loop:**  
   - Authenticates to Infrahub using configured credentials.  
   - Executes the specified query on Infrahub and retrieves metadata about the resulting artifact.  
   - Downloads the artifact (e.g., a Helm chart or Kubernetes manifest bundle).  
   - Parses and validates the artifact content.  
   - Applies the artifact manifests to the Kubernetes cluster, updating or creating resources as needed.

4. **Status Update:** Vidra updates the status section of the `InfrahubSync` resource, reporting success, errors, or pending states.

---

## 4. Design Patterns and Technologies

- **Operator SDK & Controller Runtime:** Vidra leverages the Operator SDK framework, providing scaffolding for controller creation, client interactions, and CRD management.
- **Domain-Driven Design (DDD):** Clear separation between domain entities, business logic, and infrastructure concerns to improve maintainability.
- **Interface Abstraction:** External system clients (Infrahub API, Kubernetes client) are abstracted by interfaces, facilitating unit testing and mocking.
- **Reconciliation Model:** Follows Kubernetes’ declarative reconciliation paradigm, ensuring eventual consistency between desired and actual state.

---

## 5. Error Handling and Resilience

Vidra implements comprehensive error handling strategies:

- Transient failures during API calls trigger retries with exponential backoff.
- Validation errors halt reconciliation with clear status messages.
- Unexpected errors are logged and surfaced in resource conditions, enabling operators to diagnose issues.
- Concurrency controls prevent conflicting sync operations from corrupting cluster state.

---

## 6. Extensibility and Future Directions

The modular architecture of Vidra allows:

- Easy integration of new artifact formats (e.g., Terraform, Ansible playbooks).
- Addition of new Infrahub query capabilities and authentication methods.
- Enhanced observability via metrics and tracing.

This design ensures Vidra can evolve to meet growing infrastructure automation demands and diverse deployment scenarios.

---

## Summary

This chapter has outlined the architecture of Vidra, the Infrahub operator, emphasizing its layered structure, key components, data flow, and design principles. Vidra embodies modern Kubernetes operator practices to deliver a robust and extensible solution for automated infrastructure synchronization.
