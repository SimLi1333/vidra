---
title: Testing
sidebar_position: 6
---
import Admonition from '@theme/Admonition';

# Vidra Operator Testing Strategy

This document outlines the approaches, methodologies, and types of tests that ensure the **Vidra Operator** components function as expected.

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
- **EnvTest**: Part of the `controller-runtime` library that spins up a local Kubernetes API server and etcd for testing purposes. This is particularly useful for integration and controller-level tests.
- **Kubernetes Client Libraries**: Used to interact with the Kubernetes API, allowing tests to create, update, and delete resources as needed.
- **Mockgen**: A tool for generating mock implementations of real interfaces, which can be used to isolate components during testing.
- **Fake Clients**: Used to simulate a second Kubernetes API for testing purposes, allowing us to test interactions without requiring a second envtest cluster just for the multicluster feature, while still having more realistic API interactions for normal controller interaction.

## Strategies: Test Approach

Our testing approach is divided into two primary types:

### Unit Tests

Unit tests of single functions are mainly written for adapter packages. For the controller logic, it makes more sense to use integration tests, as the controller logic is tightly coupled with the Kubernetes API and CRDs. That’s why most Vidra Operator tests are implemented using the `envtest` environment from `controller-runtime`. While `envtest` is typically used for integration testing the whole software at once, we leverage it even for isolated logic tests to provide realistic API interactions without requiring mocking or fake clients.

This approach allows us to:

- Test reconciliation logic and Kubernetes resource interactions against a real API server and etcd.
- Avoid discrepancies between fake clients and real controller behavior.
- Maintain higher confidence in test accuracy, especially when working with CRDs and status updates.

Tests remain lightweight and fast due to `envtest`’s in-memory setup, making them practical for local development and CI validation.

We created a fake mocked client to simulate failures of the Kubernetes API server, which allows us to test how the controller handles errors and retries, as there is currently no other way to simulate API server failures in a realistic way.

All `VidraResource` reconciliation tests are executed twice: once applying the resources to the local cluster and once to a simulated second cluster. This ensures that the reconciliation logic works correctly in both single-cluster and multi-cluster scenarios.

The k8s adapter packages are also tested using `envtest`, but with a focus on isolated logic rather than full controller behavior. This allows us to verify that the adapter functions correctly interact with the Kubernetes API, such as creating a second client for interacting with a different cluster.

The event-driven functionality is tested by mocking the callback functions triggered when a resource is created or updated. This approach allows us to verify that the controller responds appropriately to events and updates the `VidraResource` status as expected. Notably, this event-based test runs successfully in local environments but is currently disabled in CI due to inconsistent behavior. To prevent failures in CI, the test is conditionally skipped when the `GITHUB_ACTIONS` environment variable is set (`if syscall.Getenv("GITHUB_ACTIONS") == "true": skip`).

### End-to-End Tests

E2E tests verify that multiple components of the operator work correctly when interacting with a real (but local and temporary) Kubernetes API server. These tests use the real image built and uploaded to the GitHub registry, ensuring that the operator behaves as expected in a realistic environment.

They deploy the operator to a real Kubernetes cluster and monitor its behavior and metrics using `prometheus` and webhooks. They use `cert-manager` for creating TLS certificates for webhooks (such as admission or conversion webhooks).

The tests ensure the operator’s:

- **Installation**: Verifying that the operator can be installed and configured correctly from the online repository.
- **RBAC**: Ensuring that the operator has the necessary permissions to create and manage resources.

## Test Coverage Goals

We aim to maintain at least **80% test coverage** for the Vidra Operator. Achieving this level of coverage helps ensure:

- Stability over time
- Safe refactoring and maintenance
- Quick feedback cycles in development
- No need to test every error catch block, but rather focus on critical paths and business logic
- Rather than chasing 100% coverage, we prioritize testing business logic, like running all `VidraResource` reconciliation tests again for the multicluster feature, ensuring that the core functionality remains intact.

Coverage reports are generated and published in CI for visibility and enforcement using Codecov.

<Admonition type="note" title="Local Coverage Check">
A local visual coverage can be checked using:
```bash
make test
go tool cover -html=cover.out
```
</Admonition>
---
