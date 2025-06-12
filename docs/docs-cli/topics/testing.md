---
title: Testing the KubeCLI Adapter
position: 2
---

# Unit Tests for KubeCLI Adapter

This document describes the unit tests for the business logic and relevant functions of the KubeCLI adapter. The tests use mocking where appropriate. Cobra code and certain KubeCLI functions that interact directly with external tools are intentionally not tested.

## Test Coverage

- All business logic functions in the KubeCLI adapter are covered.
- External dependencies and system calls are mocked.
- No tests are written for Cobra CLI wiring or direct invocations of external binaries.

## Mocking

All interactions with the Kubernetes API are mocked using a suitable mocking framework (e.g., GoMock for Go, unittest.mock for Python).

## Exclusions

- Cobra CLI command parsing and execution are not tested.
- Direct execution of external tools (e.g., `kubectl`) is not tested.

---

**Note:** For implementation details, refer to the actual test files and mocking setup in the codebase.