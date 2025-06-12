---
id: home
slug: /
title: Introduction
position: 1
---

## Vidra CLI Tool

The **Vidra CLI** is a command-line utility designed to simplify the management of Vidra configurations and secrets. With the CLI, you can easily:

- Add and manage Vidra configuration files
- Create and update Infrahub credential secrets
- Manage cluster secrets securely
- Generate and apply `InfrahubSync` CRs

This tool streamlines setup and ongoing operations, making it easier to integrate Vidra into your Kubernetes workflows. For detailed usage instructions, refer to the [CLI Usage Guide](../guides/usage).

### Features
- **Cluster Management**: Easily apply, delete, and list cluster kubeconfig secrets for multicluster support.
- **Credential Management**: Generate and manage Infrahub credential secrets with ease.
- **InfrahubSync Management**: Create and manage InfrahubSync resources for syncing configurations across clusters.
- **Autocompletion**: Supports autocompletion for various shells, enhancing usability.

It automatically loads `kubeconfig` from your local context and encodes the secret data in `base64`, significantly reducing the effort required to create and manage Kubernetes secrets for Vidra.

<div align="center" style={{ marginTop: '3em' }}>
    <img src="../img/cli-help.png" alt="Vidra CLI" width="600"/>
</div>
