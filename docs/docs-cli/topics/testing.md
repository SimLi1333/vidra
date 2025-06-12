---
title: Testing the Vidra CLI
position: 2
---

# Unit Tests for the Vidra CLI

This document describes the unit tests for the business logic and relevant functions of the KubeCLI adapter. The tests use mocking where appropriate. Cobra code and certain KubeCLI functions that interact directly with external tools are intentionally not tested.

## Test Coverage

- All business logic functions in the Vidra CLI service layer.
- External dependencies and system calls are mocked.
- No tests are written for Cobra CLI wiring or direct invocations of external tools like kubectl.

## Mocking

All interactions with the KubeCLI adapter and other external dependencies are mocked using the `gomock` package. This allows us to isolate the business logic from the actual Kubernetes API and other system calls, ensuring that tests are fast and reliable.

## Exclusions

- Cobra CLI command parsing and execution are not tested, because we do not want to test the CLI framework itself.
- Direct execution of external tools (e.g., `kubectl`) is not tested.

---

**Note:** For implementation details, refer to the actual test files and mocking setup in the codebase.