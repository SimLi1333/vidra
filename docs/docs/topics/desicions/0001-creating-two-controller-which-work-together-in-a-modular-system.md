---
title: Creating Two Controllers Which Work Together in a Modular System
sidebar_position: 2
---

# Creating Two Controllers Which Work Together in a Modular System

## Context and Problem Statement

As part of the Vidra project, we are developing a Kubernetes Operator to synchronize external resources from Infrahub into Kubernetes. While the Operator’s primary responsibility is to manage `InfrahubSync` custom resources, we identified a need for additional processing and reconciliation of downstream `VidraResource` objects.

Initially, we considered implementing a single controller to handle both the synchronization of external artifacts and the lifecycle of the resulting Kubernetes resources. However, this approach led to tight coupling of responsibilities, limiting extensibility, making testing harder, and increasing cognitive overhead.

To address these concerns, we explored whether splitting the responsibilities into two distinct controllers would lead to a more modular and maintainable architecture.

## Considered Options

* **One controller**  
  A single controller that handles both the fetching and transformation of Infrahub data and the reconciliation of downstream resources.

* **Two controllers**  
  One controller (`InfrahubSyncReconciler`) handles synchronization from Infrahub, and a second controller (`VidraResourceReconciler`) manages the lifecycle of resulting Kubernetes resources.

## Decision Outcome

**Chosen option: "Two controllers"**, because:

- It encourages **separation of concerns** by isolating synchronization logic from resource reconciliation.
- It allows each controller to be **tested independently**, improving reliability and development speed.
- It supports **scalability and future extensibility**, making it easier to add features such as validation, transformation pipelines, or reconciling multiple resource types.
- It aligns with Kubernetes Operator best practices, where each controller focuses on a specific resource kind.
- It makes the system **more maintainable**, as responsibilities are clearer and better encapsulated.

### Consequences

* Good, because it leads to a **cleaner and more modular design**, reducing coupling and improving testability.
* Bad, because it introduces **slightly more initial complexity**, requiring coordination between the two controllers and managing shared state or events.

Overall, we believe using two controllers strikes the right balance between clarity, flexibility, and maintainability for Vidra’s architecture.
