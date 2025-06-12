import Admonition from '@theme/Admonition';

# Vidra CLI Architecture

The Vidra CLI is structured following clean architecture principles, using [Go Cobra](https://github.com/spf13/cobra) for command-line parsing and execution.

## Folder Structure

- [**`cmd/`**](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli/cmd)  
    Contains all CLI commands and operations. Each command is defined as a separate file or package, leveraging Cobra for argument parsing and command handling.

- [**`internal/service/`**](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli/internal/service)  
    Implements the core logic for each command. This layer contains business logic and some tests, ensuring that command implementations remain decoupled from infrastructure concerns.

- [**`internal/adapter/`**](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli/internal/adapter)  
    Provides reusable Kubernetes client (`kubecli`) functions. This adapter abstracts infrastructure details and is accessed via interfaces, enabling easy mocking and testing.

## Architectural Overview

- **Commands** in `cmd/` delegate to **services** in `internal/service/`.
- **Services** interact with **adapters** in `internal/adapter/` for Kubernetes operations.
- **Adapters** encapsulate external dependencies, promoting testability and maintainability, and enable reuse of the same function for all commands/services.

This separation ensures a modular, testable, and maintainable CLI codebase.

<Admonition type="note" title="Note">
You can explore the implementation of these layers in the [vidra-cli directory](https://github.com/infrahub-operator/vidra/tree/main/vidra-cli).
</Admonition>

## kubecli Adapter

The `kubecli` adapter provides a set of reusable functions for interacting with Kubernetes clusters. It abstracts the complexity of Kubernetes API calls, allowing reusable operations across different commands. This adapter is designed to be used by the service layer, providing a consistent interface for Kubernetes operations.

## Resource Naming Strategy

To generate meaningful yet unique resource names, the CLI uses hashing of unique input values. For example, when creating a kubeconfig for a cluster, the resource name is derived as follows:

```
cluster-kubeconfig-{hash of cluster name}
```

This approach ensures that resource names are both descriptive and collision-resistant, while avoiding exposure of sensitive or overly long identifiers.

## Error Handling

Each CLI command has its own error handler that is responsible for formatting and returning errors in a consistent manner. This allows for better user experience and easier debugging.

For example, if a cluster secret apply fails because the command is incomplete, the error handler will also prompt the user with the available cluster contexts in their kubeconfig file, allowing them to select one to create a kubeconfig secret for that cluster.

Default timeouts for all commands are set to 1 minute using a context deadline. This ensures that commands do not hang indefinitely and provides a consistent timeout behavior across the CLI.

## Function Reuse

The CLI is designed to maximize function reuse across different commands. For example, the `kubecli` adapter provides functions for applying, deleting, and retrieving Kubernetes secrets, which are used by multiple commands. This reduces code duplication and ensures consistency in how Kubernetes operations are performed.

Additionally, it adds maintainability and testability, as changes to the underlying logic only need to be made in one place.
