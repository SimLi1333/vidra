---
title: Decision to Apply Clean Architecture in Vidra Operator
sidebar_position: 8
---

# Decision to Apply Clean Architecture in Vidra Operator

## Context and Problem Statement

To ensure maintainability, testability, and a clear separation of concerns, we needed to decide on the architectural approach for the Vidra Operator. The main challenge was structuring the codebase to support evolving requirements, integration with Kubernetes, and external systems like Infrahub, while keeping the core business logic independent.

## Considered Options

* **Monolithic structure**  
    All logic (Kubernetes, domain, external integrations) is mixed in the Reconcile function, leading to tight coupling and making testing harder. (Many operators follow this pattern.)

* **Clean/Onion Architecture**  
    Separate the codebase into:
    - **Controller**: Handles direct Kubernetes logic (reconcilers, API interactions).
    - **Domain**: Contains core types, business logic, and interfaces, independent of infrastructure.
    - **Adapters**: Implement interfaces for external systems (Infrahub, Kubernetes shared utilities).

## Decision Outcome

**Chosen option: "Clean/Onion Architecture"**, because it enables a clear separation between business logic and infrastructure concerns. The domain layer remains decoupled from Kubernetes and Infrahub specifics, making the codebase easier to test, extend, and maintain.

### Consequences

* Positive: Improves modularity, testability, and adaptability to future changes or integrations.
* Negative: Introduces some initial complexity and requires discipline to maintain boundaries between layers.

## References

- Robert C. Martin, *Clean Architecture: A Craftsman's Guide to Software Structure and Design*, Prentice Hall, 2017.