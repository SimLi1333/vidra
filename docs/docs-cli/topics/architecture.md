# Vidra CLI Architecture

The Vidra CLI is structured following clean architecture principles, using [Go Cobra](https://github.com/spf13/cobra) for command-line parsing and execution.

## Folder Structure

- **`cmd/`**  
    Contains all CLI commands and operations. Each command is defined as a separate file or package, leveraging Cobra for argument parsing and command handling.

- **`internal/service/`**  
    Implements the core logic for each command. This layer contains business logic and some tests, ensuring that command implementations remain decoupled from infrastructure concerns.

- **`internal/adapter/`**  
    Provides reusable Kubernetes client (`kubecli`) functions. These adapters abstract infrastructure details and are accessed via interfaces, enabling easy mocking and testing.

## Architectural Overview

- **Commands** in `cmd/` delegate to **services** in `internal/service/`.
- **Services** interact with **adapters** in `internal/adapter/` for Kubernetes operations.
- **Adapters** encapsulate external dependencies, promoting testability and maintainability.

This separation ensures a modular, testable, and maintainable CLI codebase.