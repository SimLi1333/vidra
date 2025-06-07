---
title: Creating Two Controllers for Modular System Design
sidebar_position: 2
---

# Creating Two Controllers for Modular System Design

## Context and Problem Statement

We needed to decide whether to implement a single controller to handle both synchronization with Infrahub and Kubernetes managed resource lifecycle management (continuous delivery), or to split these responsibilities into two dedicated controllers.

A single controller can simplify deployment but may become complex and harder to maintain as responsibilities grow. Our use case involves both integrating with Infrahub (external system) and managing the lifecycle of Kubernetes resources, which are distinct concerns.

## Considered Options

* **Single controller handling both Infrahub sync and resource lifecycle**  
  Simpler deployment, but risks mixing concerns and increasing code complexity.

* **Two controllers: one for Infrahub sync, one for resource lifecycle management**  
  Clear separation of concerns, easier to test and maintain, but introduces more components to deploy and coordinate.

## Decision Outcome

**Chosen option: "Two controllers: one for Infrahub sync, one for resource lifecycle management"**, because:

- Separation of concerns leads to cleaner, more maintainable codebases.
- Each controller can evolve independently, allowing for focused improvements and optimizations.
- Issues in one controller (e.g., Infrahub connectivity) do not directly impact the other (resource lifecycle).
- Testing and debugging are simplified due to reduced scope per controller.

### Consequences

* Good, because it provides **modularity, maintainability**, and aligns with **best practices for controller design** in Kubernetes.
* Bad, because it introduces **additional deployment complexity** and requires **coordination between controllers**.
