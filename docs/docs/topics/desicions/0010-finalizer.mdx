---
title: Use of Finalizer for Resource Cleanup
sidebar_position: 11
---

# Use of Finalizer for Resource Cleanup

## Context and Problem Statement

Previously, we relied solely on Kubernetes ownership (`ownerReferences`) for managing the lifecycle of resources created by Vidra. This approach had limitations, especially in multi-cluster scenarios and when ensuring proper cleanup of all managed resources.
## Considered Options

* **Rely solely on Kubernetes `ownerReferences`**  
    Use the built-in Kubernetes mechanism to automatically clean up dependent resources when the parent is deleted. This approach is simple but limited in multi-cluster scenarios and may not guarantee cleanup of all resources.

* **Implement a custom finalizer**  
    Attach a finalizer to Vidra resources, ensuring that all managed resources are explicitly deleted before the Vidra resource itself is removed. This provides more control and reliability, especially in complex or multi-cluster environments.

## Decision

**Chosen option: "Implement a custom finalizer"**, because it allows us to ensure that all resources created and managed by Vidra are properly cleaned up before the Vidra resource itself is deleted. This is particularly important in multi-cluster scenarios where Kubernetes ownership is not possble.

## Consequences

* Good, because it provides a reliable mechanism for resource cleanup, ensuring that no resources are left orphaned after a Vidra resource is deleted.
* Good, because it allows using the same cleanup logic across different clusters, ensuring consistency in resource management.
* Bad, because it introduces additional complexity in managing finalizers and requires careful handling to avoid resource leaks if the finalizer logic fails.
