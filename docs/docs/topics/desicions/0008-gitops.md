---
title: Adopting GitOps Principles for Vidra Operator
sidebar_position: 9
---

# Adopting GitOps Principles for Vidra Operator

## Context and Problem Statement

We decided to align our operational and deployment practices with the [four principles of GitOps](https://opengitops.dev). This decision aims to improve automation, reliability, and transparency in managing the Vidra Operator and its resources.

## Considered Options

* **Continue with ad-hoc/manual operations**  
    Operations are performed manually or with custom scripts, leading to inconsistencies and reduced auditability.

* **Adopt GitOps principles**  
    Operations and desired state are managed declaratively in Infrahub, with automated reconciliation and strong audit trails.

## Decision Outcome

**Chosen option: "Adopt GitOps principles"**, because it enables us to manage infrastructure and application state declaratively, automate deployments, and ensure that the cluster state matches the source of truth in Infrahub.

### The Four GitOps Principles

1. **Declarative Descriptions**  
     A system managed by GitOps must have its desired state expressed declaratively.

2. **Versioned and Immutable Storage**  
     Desired state is stored in a way that enforces immutability, versioning and retains a complete version history.

3. **Automated Reconciliation**  
    Software agents automatically pull the desired state declarations from the source.

4. **Continuous Operations**  
     Software agents continuously observe actual system state and attempt to apply the desired state.

### Consequences

* Good, because it increases reliability, auditability, and repeatability of operations.

