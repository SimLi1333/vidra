---
title: Transitioning Vidra Operator to Cluster Scope
sidebar_position: 4
---

# Transitioning Vidra Operator to Cluster Scope

## Context and Problem Statement

We needed to decide whether the Controller should be namespace-scoped or cluster-scoped.

While namespace-scoped Operators offer isolation, they are limited as they can not take ownership of resources in other namespaces. Our use case requires managing resources across multiple namespaces and potentially work with ownership could improve this.

## Considered Options

* **Namespace-scoped** controller and CRDs
    Limited to managing resources within a single namespace, may require additional logic to clean up resources using finalizers.

* **Cluster-scoped** controller and CRD
    Can manage resources across all namespaces and utilize Kubernetes ownership features effectively.

## Decision Outcome

**Chosen option: "cluster-scoped"**, because While namespace-scoped is possible especaly in combination with finalizers to delete resources again, we do not want to lose the benefits of kubernetes ownership. We chose to go with cluster-scoped and leave us all options open to use ownership or finalizers to remove managed resources again.

### Consequences
* Good, because it provides flexibility, better visibility, and aligns with ownership and management patterns in Kubernetes.
* Bad, because it requires additional permissions on the cluster and potentially access control to ensure security.