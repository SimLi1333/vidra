---
title: Namespace Handling for Vidra-Managed Resources
sidebar_position: 10
---

# Namespace Handling for Vidra-Managed Resources

## Context and Problem Statement

As part of designing the Vidra Operator, we needed to decide how to handle the namespaces of resources managed by Vidra. Specifically, we considered whether each managed resource should have a distinct namespace, or if the namespace should be specified and handled within the `InfrahubSync` Custom Resource Definition (CRD).

This decision impacts how resources are organized, how access control is managed, and how flexible the operator is in supporting different deployment scenarios.

## Considered Options

* **Distinct Namespace per Managed Resource:**  
  In this approach, each resource managed by Vidra is deployed into its own dedicated namespace. This provides strong isolation between resources, which can enhance security and reduce the risk of unintended interactions. However, it also increases the complexity of namespace management and makes resource discovery more challenging, as administrators must track and manage a larger number of namespaces.

* **Namespace Specified in InfrahubSync CRD:**  
  The namespace for each managed resource is defined within the `InfrahubSync` CRD. This approach enables grouping multiple resources in a single namespace or distributing them as needed. It implies that all resources in one artifact group (like Webserver) would be in one namespace. It simplifies management and aligns with Kubernetes patterns for multi-tenancy and resource grouping.

## Decision Outcome

**Chosen option: both**, because we want to support as many deployment scenarios as possible. The operator will use the namespace defined in the `InfrahubSync` CRD as a fallback if the resources do not have a namespace. This approach maximizes flexibility, allowing users to set a default namespace at the CRD level while still enabling per-resource overrides when needed.

### Consequences

* Good, because it enables flexible resource organization, supports multi-tenancy, and reduces namespace sprawl.
* Bad, because it requires careful validation and documentation to ensure users specify namespaces correctly and understand the implications for access control and resource isolation.