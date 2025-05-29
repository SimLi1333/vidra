---
title: Transitioning Vidra Operator to Cluster Scope
sidebar_position: 3
---

# Transitioning Vidra Operator to Cluster Scope

## Context and Problem Statement

We needed to decide whether the Controller should be namespace-scoped or cluster-scoped.

While namespace-scoped Operators offer isolation, they are limited as they cannot take ownership of resources in other namespaces. Our use case requires managing resources across multiple namespaces, and leveraging ownership could improve resource lifecycle management.

## Considered Options

* **Namespace-scoped controller and CRDs**  
  Limited to managing resources within a single namespace, may require additional logic to clean up resources using finalizers.

* **Cluster-scoped controller and CRD**  
  Can manage resources across all namespaces and utilize Kubernetes ownership features effectively.

## Decision Outcome

**Chosen option: "Cluster-scoped controller and CRD"**, because:

- While a namespace-scoped controller is technically feasible—especially when combined with finalizers to handle cleanup—we want to fully benefit from Kubernetes ownership semantics.
- A cluster-scoped design gives us the flexibility to manage multi-namespace resources cleanly and consistently.
- It allows us to support advanced use cases and infrastructure-level resources in the future.
- We leave the door open to use either finalizers or ownership patterns for resource cleanup, depending on what best fits the use case.

### Consequences

* Good, because it provides **flexibility, better visibility**, and aligns with **ownership and management patterns** in Kubernetes.
* Bad, because it may require **additional access control** and careful **RBAC management** to ensure security.
