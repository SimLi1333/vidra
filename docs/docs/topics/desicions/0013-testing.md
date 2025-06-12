---
title: Testing Framework Decision
sidebar_position: 14
---

# Testing Framework

## Context and Problem Statement

Every real project needs to be tested. There are multiple testing libraries and frameworks to test Go applications, which can be used in addition to the default Go testing library. For our controllers (`infrahubsync` and `vidraresource`) and adapter, we required a framework that supports comprehensive unit testing and integration-style tests, as our tests need to cover the entire reconciliation or setup process of the controller (not just single functions, but the whole logic block to bring the tests to a higher level). Clean code is maintained by splitting logic into multiple functions, all exercised by these tests.

## Considered Options

* **Testify**  
    A popular testing toolkit for Go that provides assertions, mocking, and more. However, it is primarily focused on unit testing and does not provide a structured way to write integration tests.

* **Ginkgo/Gomega**  
    A BDD-style (Behavior-driven development) testing framework that supports both unit and integration tests. It provides a rich set of assertions and allows for structured test organization, making it suitable for complex integration tests.

* **GoSpec**  
    A lightweight BDD-style testing framework, but it lacks some of the advanced features and community support of Ginkgo/Gomega.

* **GoConvey**  
    A testing framework that provides a web UI for test results and supports BDD-style tests. However, it is less commonly used in the Kubernetes community and does not integrate as well with the Operator SDK.

## Decision Outcome

**Chosen option: Ginkgo/Gomega**, because it is widely used in the Kubernetes community to test Kubernetes operators. It is also used in the Operator SDK project, which is leveraged by the Vidra Operator. Additionally, tests written with Ginkgo/Gomega are easy to read and understand.

To facilitate testing, we used `envtest` to provide a real Kubernetes API server environment. For dependency injection and mocking, we used `mockgen` and custom-written mocks, especially for the adapter layer. The factory pattern was applied in the multicluster and event-based Kubernetes code to enable flexible test setups.

End-to-end (E2E) tests are performed using Cert-Manager and Prometheus. These tests deploy the Vidra Operator from the container registry into a real Kubernetes cluster, then monitor the operator's startup and behavior. This approach ensures that the operator works as expected in a production-like environment, validating integration with external components and overall system reliability.

### Consequences

* Good, because it aligns with community standards, improves readability, and supports integration-style testing.
* Bad, because it introduces an additional dependency and requires learning the Ginkgo/Gomega syntax.