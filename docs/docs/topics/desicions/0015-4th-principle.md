---
title: Implementing the 4th GitOps Principle in Vidra Operator
sidebar_position: 16
---

# Implementing the 4th GitOps Principle in Vidra Operator

## Context and Problem Statement

The 4th GitOps principle states:  
**"Software agents continuously observe actual system state and attempt to apply the desired state."**

During implementation, we realized that `VidraResource` reconciliation must also run in a loop to ensure the actual state matches the desired state.

## Considered Options

* **Time-based reconciliation loop**  
    Implement a simple loop that periodically checks the state of managed resources and reconciles them if they differ from the desired state. This is straightforward and ensures regular updates.

* **Event-based reconciliation**  
    Trigger reconciliation immediately when relevant changes are detected in managed resources. This provides rapid response to changes but may be more resource intensive.

## Decision Outcome

**Chosen option: Both time-based and event-based reconciliation.** because it combines the reliability of periodic checks with the responsiveness of event-driven updates and allows the user to configure the reconciliation interval how he wants.

We first implemented a simple, time-based reconciliation loop using the `reconcileAfter` return value. The interval is configurable via Vidra `ConfigMap`, allowing operators to tune how often reconciliation occurs.

We then added event-based reconciliation, which triggers reconciliation immediately when relevant changes are detected in managed resources. This enables faster reaction to changes but may be more resource intensive.

### Consequences

* Good, because it combines reliability (periodic checks) with responsiveness (event-driven updates).
* Good, because it allows operators to configure the reconciliation interval based on their specific needs and resource constraints. (eventbased can be activated even just for one `InfrahubSync` resource)
* Bad, because event-based reconciliation can be more resource intensive, so careful configuration and monitoring are required.
* Bad, because implementing both mechanisms increases complexity, requiring careful management of reconciliation logic to avoid conflicts or redundant operations.

