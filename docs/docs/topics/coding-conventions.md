---
title: Coding Conventions
sidebar_position: 4
---
# Coding Conventions

We follow the [**Effective Go**](https://golang.org/doc/effective_go.html) guidelines as the foundation for our coding conventions. This document summarizes the most important practices and tools enforced in this project.

---

## Naming

### General Rules

- **camelCase** for:
  - Local variables
  - Private functions and methods (unexported)
- **PascalCase** for:
  - Exported functions
  - Types (structs, interfaces, etc.)
  - Constants that need to be exported

### Examples

| Context                | Example                      |
|------------------------|------------------------------|
| Variable name          | `infrahubSync`               |
| Exported function name | `GetSortedListByLabel()`     |
| Exported type          | `DynamicMulticlusterFactory` |
| Unexported function    | `processArtifacts()`         |

---

## Formatting

All code must be formatted using the official Go formatting tool:

- [`gofmt`](https://pkg.go.dev/cmd/gofmt): Standard formatting tool from the Go toolchain.
  - Most editors apply this automatically.
  - Run `gofmt -s -w .` to format code and simplify syntax.

---

## Linting

We use [`golangci-lint`](https://golangci-lint.run/) as the primary linting tool. It aggregates multiple linters, including `staticcheck`, `govet`, `gocritic`, and others.

### Usage

To run linting locally:

```bash
make lint
```

---

## Architecture

Use interfaces to define behavior and decouple components. This allows for easier testing and mocking.

### Example

```go
// Domain Layer
type DynamicMulticlusterFactory interface {
  GetCachedClientFor(ctx context.Context, serverURL string, k8sClient client.Client) (client.Client, error)
}
// Adapter Layer
type DynamicMulticlusterFactory struct {
  mu      sync.Mutex
  clients map[string]client.Client
}

func NewDynamicMulticlusterFactory() *DynamicMulticlusterFactory {
  return &DynamicMulticlusterFactory{
    clients: make(map[string]client.Client),
  }
}

func (f *DynamicMulticlusterFactory) GetCachedClientFor(ctx context.Context, serverURL string, k8sClient client.Client) (client.Client, error) {
    // Implementation details...
}
```