---
title: Managed Resource Ownership and Cleanup Labels
sidebar_position: 11
---

# Managed Resource Ownership and Cleanup Labels

## Context and Problem Statement

We encountered a problem where a managed resource could be owned by two different `InfrahubSync` resources. This created ambiguity in resource ownership and cleanup responsibilities.

To address this, we needed a way to track both general management by Vidra and specific ownership by one or more Infrahub resources. Additionally, we wanted to avoid overwriting existing resources in the cluster unless they are explicitly managed by Vidra.

## Considered Options

* **No explicit ownership tracking**  
    Relies on Kubernetes owner references or manual tracking, which can lead to conflicts and unclear cleanup responsibilities.

* **Implement labels for management and ownership**  
    Use a `managed-by` label to indicate Vidra management, and an `owned-by` label containing a list of `VidraResources` responsible for the managed resource.

## Decision Outcome

**Chosen option: "Implement managed-by and owned-by labels"**  
We decided to add two labels to all managed resources:
- `managed-by`: Indicates the resource is managed by Vidra.
- `owned-by`: Contains a list of `VidraResources` that currently own the resource.

This approach enables:
- Accurate tracking of which `VidraResources` are responsible for cleanup.
- Additional benefit: Prevents overwriting resources not managed by Vidra by checking the `managed-by` label before creating or updating resources.

* Good, because it enables correct cleanup, prevents accidental overwrites, and clarifies resource ownership.
* Bad, because it introduces additional label management logic and requires consistent label updates as ownership changes.