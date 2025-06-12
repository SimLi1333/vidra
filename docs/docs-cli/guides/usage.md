---
title: Usage
sidebar_position: 2
---
import Admonition from '@theme/Admonition';

# Usage

`vidra-cli [command]`

vidra-cli is a command-line tool for managing the vidra-cli Operator and related resources.

## Available Commands

- **cluster**: Manage clusters for multicluster vidra-cli Operator.
- **completion**: Generate the autocompletion script for the specified shell.
- **config**: Configure the vidra-cli Operator.
- **credentials**: Manage Infrahub credential Secrets.
- **help**: Display help information about any command.
- **infrahubsync**: Manage InfrahubSync resources.

## Flags

- `-h`, `--help`: Show help for the `vidra` command.

## Operations

Each command has the same set of operations:

### Example credentials command

- **apply**: Generate and apply an Infrahub credentials secret to Kubernetes.
- **delete**: Remove an Infrahub credentials secret by URL.
- **get**: Retrieve an Infrahub credentials secret by URL.
- **list**: List all Infrahub credentials secrets.

## Examples

```sh
# Apply a kubeconfig secret for a cluster, reading from your kubeconfig file
vidra-cli cluster apply admin@ba-iac -n secrets

# Delete a kubeconfig secret for a cluster
vidra-cli cluster delete admin@ba-iac -n secrets

# Get a kubeconfig secret for a cluster
vidra-cli cluster get --name my-cluster

# List all cluster kubeconfig secrets
vidra-cli cluster list
```
<Admonition type="note" title="Note">
Running `vidra-cli cluster apply` (incomplete command) will prompt you with the available cluster contexts in your kubeconfig file. You can select one of them to create a kubeconfig secret for that cluster.
</Admonition>

```sh
# Apply an Infrahub credentials secret
vidra-cli credentials apply https://infrahub.example.com --username admin --password secret

# Delete an Infrahub credentials secret by URL
vidra-cli credentials delete https://infrahub.example.com

# Get an Infrahub credentials secret by URL
vidra-cli credentials get https://infrahub.example.com

# List all Infrahub credentials secrets
vidra-cli credentials list
```
```sh
# Apply a vidra-cli configuration
vidra-cli config apply --query-name ArtifactIDs -r 5m -s 1m

# Delete a vidra-cli configuration
vidra-cli config delete

# Get a vidra-cli configuration from the default namespace
vidra-cli config get -n default

# List all vidra-cli configurations 
vidra-cli config list 
```

Apply an `InfrahubSync` resource:
```sh
vidra-cli infrahubsync apply "http://198.19.248.5:8000" -a Webserver_Manifest -b main2 -d 2025-04-09T00:00:00Z -s https://kubernetes.default.svc -N default -e
```
<Admonition type="note" title="Note">
Please use the -h flag to get more information about each command and its options.
</Admonition>
