---
title: Storing Resource State
sidebar_position: 13
---
import Admonition from '@theme/Admonition';

# Storing Resource State

## Context and Problem Statement

We want to persist the state of resources managed by the Vidra Operator, specifically tracking the last successful reconcile, last error, and the current Sync/Deploy state for both `InfrahubSync` and `VidraResource` objects. This information is stored in the `status` section of each resource.

<Admonition type="note" title="Note">
This was not considered during planning. If there had been no time to implement it, we would have left it out.
</Admonition>

## Considered Options

* **Do not store state information:**  
    Do not persist the resource's state, last successful reconcile, or last error in the `status` section. This keeps the resource definition simple but loses valuable operational insights and troubleshooting data.

* **Store state information in status:**  
    Persist the state, last successful reconcile, and last error in the `status` section of each resource. This provides visibility into resource health and operations but introduces the challenge of status updates potentially triggering unnecessary reconcile loops.

## Decision Outcome

**Chosen option: "Store state information in status"**, because it allows us to track the operational state of resources effectively, providing valuable insights into their health and history. This is particularly useful for monitoring and troubleshooting purposes.

We decided to store the state, last successful reconcile, and last error in the `status` section of both `InfrahubSync` and `VidraResource` objects. This allows us to track the operational state of resources effectively.

A challenge arose: updating the status (e.g., setting the state to "Running" or adding a timestamp) would normally trigger another reconcile loop, potentially causing a reconciliation loop.
To address this, we implemented a commonly used predicate that prevents reconciles when only the `status` or `metadata` section changes (`GenerationChangedPredicate{}`). This avoids infinite loops while still allowing us to update the status as needed.

### Consequences
* Good, because it provides visibility into the operational state of resources, enabling better monitoring and troubleshooting.
* Bad, because it requires careful handling of status updates to avoid triggering unnecessary reconcile loops.