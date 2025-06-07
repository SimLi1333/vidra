import Admonition from '@theme/Admonition';

# Vidra CLI Architecture

The Vidra CLI is structured following clean architecture principles, using [Go Cobra](https://github.com/spf13/cobra) for command-line parsing and execution.

## Folder Structure

- [**`cmd/`**](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli/cmd)
    Contains all CLI commands and operations. Each command is defined as a separate file or package, leveraging Cobra for argument parsing and command handling.

- [**`internal/service/`**](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli/internal/service)  
    Implements the core logic for each command. This layer contains business logic and some tests, ensuring that command implementations remain decoupled from infrastructure concerns.

- [**`internal/adapter/`**](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli/internal/adapter)
    Provides reusable Kubernetes client (`kubecli`) functions. This adapter abstract infrastructure details and are accessed via interfaces, enabling easy mocking and testing.

## Architectural Overview

- **Commands** in `cmd/` delegate to **services** in `internal/service/`.
- **Services** interact with **adapters** in `internal/adapter/` for Kubernetes operations.
- **Adapters** encapsulate external dependencies, promoting testability and maintainability and enable reuse of the same fuction for all commands / services.

This separation ensures a modular, testable, and maintainable CLI codebase.

<Admonition type="note" title="Note">
You can explore the implementation of these layers in the [vidra-cli directory](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli).
</Admonition>