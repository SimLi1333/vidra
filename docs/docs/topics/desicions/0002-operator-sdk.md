---
title: Operator-SDK
sidebar_position: 3
---

# Operator-SDK

## Context and Problem Statement

Developing Kubernetes operators is simplified by using a dedicated SDK. An SDK abstracts much of the boilerplate and complexity, enabling faster and more reliable operator development. Several SDKs exist, each with different language support and features. Our requirement is an SDK that is robust, actively maintained, and supports Go.

## Considered Options

* **Operator-SDK (Kubebuilder wrapper)**
  First-class integration with Kubernetes APIs, Supports CRD generation, scaffolding, scorecard testing, metrics, and Helm/Ansible reuse. Used by many production-grade operators (e.g., Prometheus Operator)

* **KubeBuilder**
    A framework for building Kubernetes APIs using CRDs and controllers. It provides a solid foundation but lacks some of the advanced features of Operator-SDK.
* **Metacontroller**
    A framework for building Kubernetes controllers that can manage multiple resources. It is more focused on composition and less on full operator capabilities.
* **Daisy**
    A lightweight framework for building Kubernetes controllers in YAML, but it lacks the comprehensive features and tooling of Operator-SDK.
    * **KOPF (Kubernetes Operator Python Framework)**
    A Python-based framework for building Kubernetes operators. While we have more experience with Python, it does not align with our Go-based development stack.

## Decision Outcome

**Chosen option: "Cluster-scoped controller and CRD"**, because:
- Operator-SDK provides a comprehensive set of tools for building, testing, and deploying operators.
- It has strong community support and is widely adopted in the Kubernetes ecosystem.
- It integrates seamlessly with Go, which we need to create a robust continuous delivery operator for Infrahub.

### Consequences

* Good, because it provides **flexibility, better visibility**, and aligns with **ownership and management patterns** in Kubernetes.
* Bad, because not all teammembers are familiar with Go, but since there is more to this projet than just the operator, we can learn it together and split the work between team members with different language skills.

## Links

- [Operator-SDK Documentation](https://sdk.operatorframework.io/)
- [Kubernetes: Tools to build an operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)