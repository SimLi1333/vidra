---
title: Testing
sidebar_position: 6
---
import Admonition from '@theme/Admonition';

# Vidra Operator Testing Strategy

## Test Concept

This document outlines the approaches, methodologies, and types of tests that ensure that the **Vidra Operator** components are functioning as expected.

The primary goal of testing is to maintain a high level of code reliability and confidence when integrating changes. We aim to build a robust test suite that supports both fast local iteration and CI-based validation.

## Test Categories

The current focus of our testing strategy is on:

- **Functionality**: Ensuring that the Vidra Operator behaves correctly under different conditions.
- **Logic**: Verifying that business logic and reconciliation workflows operate as expected.

Future test categories may include:

- **Security Testing**
- **Performance and Load Testing**

## Tools

We use the following tools and frameworks to test the Vidra Operator:

- **Go `testing` package**: The standard Go testing library used as a foundation for automated test cases.
- **Ginkgo**: A BDD-style Go testing framework that enables expressive and structured test suites.
- **Gomega**: A matcher/assertion library used with Ginkgo to write clear and readable expectations in tests.
- **EnvTest**: A part of the `controller-runtime` library that spins up a local Kubernetes API server and etcd for testing purposes. This is particularly useful for integration and controller-level tests.
- **Kubernetes Client Libraries**: Used to interact with the Kubernetes API, allowing tests to create, update, and delete resources as needed.
- **Mockgen**: A tool for generating mock implementations of interfaces, which can be used to isolate components during testing.
- **Fake Clients**: Used to simulate a second Kubernetes API for testing purposes, allowing us to test interactions without requiring a second envtest cluster just for the multicluster feature.

More details on the rationale behind tool selection can be found in the internal [Testing Framework Decision] document (link TBD).

## Strategies: Test Approach

Our testing approach is divided into two primary types:

### Unit Tests

Unit tests in the Vidra Operator are implemented using the `envtest` environment from `controller-runtime`. While `envtest` is typically used for integration testing, we leverage it even for isolated logic tests to provide realistic API interactions without requiring mocking or fake clients.

This approach allows us to:

- Test reconciliation logic and Kubernetes resource interactions against a real API server and etcd.
- Avoid discrepancies between fake clients and real controller behavior.
- Maintain higher confidence in test accuracy, especially when working with CRDs and status updates.

Tests remain lightweight and fast due to `envtest`’s in-memory setup, making them practical for local development and CI validation.

### Integration Tests

Integration tests verify that multiple components of the operator work correctly when interacting with a real (but local and temporary) Kubernetes API server. These tests use:

- **EnvTest**, which provides a lightweight Kubernetes environment via embedded etcd and API server.
- Real CRs and controller managers to closely mimic production behavior.
- Gomega matchers for clear, expressive assertions.

Integration tests are typically written using Ginkgo and structured to simulate realistic lifecycle operations, such as:

- Creating and reconciling custom resources
- Observing status updates
- Validating interactions between controllers and the Kubernetes API

We recommend running integration tests as part of CI pipelines or local smoke tests.

## Test Coverage Goals

We aim to maintain at least **80% test coverage** for the Vidra Operator. Achieving this level of coverage helps ensure:

- Stability over time
- Safe refactoring and maintenance
- Quick feedback cycles in development
- No need to test every error catch block, but rather focus on critical paths and business logic
- Rather than chacing 100% coverage, we prioritize testing business logic like runing all `VidraResource` reconciliation tests again for the multicluster feature, ensuring that the core functionality remains intact.

Coverage reports are generated and published in CI for visibility and enforcement.


<Admonition type="note" title="Local Coverage Check">
A local visual coverage can be checked using:
```bash
make test
go tool cover -html=cover.out
```
</Admonition>
---

