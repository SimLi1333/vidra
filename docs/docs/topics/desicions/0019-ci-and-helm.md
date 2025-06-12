---
title: CI/CD and Helm Chart Integration for Vidra Operator
sidebar_position: 20
---

# CI/CD and Helm Chart Integration for Vidra Operator

## Context and Problem Statement

We needed to decide how to build, test, and distribute the Vidra Operator. The main options were to follow the Operator SDK's standard approach (building and publishing the operator bundle) or to provide a Helm chart for installation flexibility.

## Considered Options

* **Operator Lifecycle Management (OLM)**  
    Publish the operator bundle using the Operator SDK's recommended workflow.

* **Helm Chart**  
    In addition to the standard bundle, generate a Helm chart (using `helmify` in the `Makefile`), and publish both the bundle and the chart as GitHub packages and release assets.

## Decision Outcome

**Chosen option: "Operator Lifecycle Management (OLM) + Helm Chart"**, because it provides greater flexibility for users to install the operator in different environments (e.g., using `kubectl` or Helm). The Helm chart simplifies installation and management for users who prefer Helm (the majority), while the operator bundle ensures compatibility with Operator Lifecycle Managers (OLMs) and Kubernetes-native installations.

### Consequences

* Good, because it increases installation flexibility, making it easier for users to adopt the Vidra Operator in various environments. The Helm chart provides a familiar interface for many Kubernetes users.
* Good, because it allows users to choose their preferred installation method, whether through OLM or Helm, without forcing them into a specific workflow.
* Bad, because it adds complexity to the CI/CD pipeline and requires maintaining multiple distribution artifacts.