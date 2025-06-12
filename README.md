# Vidra
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes/kubernetes)](https://goreportcard.com/report/https://github.com/infrahub-operator/vidra) 
[![GoDoc](https://pkg.go.dev/badge/github.com/infrahub-operator/vidra)](https://pkg.go.dev/github.com/infrahub-operator/vidra)
![GitHub Release](https://img.shields.io/github/v/release/infrahub-operator/vidra?include_prereleases&sort=semver)
![Test Status](https://img.shields.io/github/actions/workflow/status/infrahub-operator/vidra/main.yaml?label=Tests)
[![codecov](https://codecov.io/github/infrahub-operator/vidra/graph/badge.svg?token=0JCAWNOMBL)](https://codecov.io/github/infrahub-operator/vidra)
![GitHub](https://img.shields.io/github/license/infrahub-operator/vidra)
<img src=".github/logo.png" alt="Nornir Conditional Runner Logo" height="200" align="right">

A continuous deployment Kubernetes operator for [Infrahub](https://github.com/opsmill/infrahub) 

`Vidra` is a Kubernetes operator designed to automate the deployment and lifecycle management of k8s resources defined in [Infrahub](https://github.com/opsmill/infrahub) on Kubernetes clusters. It streamlines continuous deployment workflows, reduces operational overhead, and simplifies the management of complex k8s deployments.

It automates syncing infrastructure artifacts from Infrahub into Kubernetes CRs. And reconciles the synced manifest with the state of the cluster, enabling continuous deployment workflows.

## Getting Started
To get started, please refer to the [installation guide](https://infrahub-operator.github.io/vidra/guides/install) and explore the [usage guide](https://infrahub-operator.github.io/vidra/guides/usage) and [API references](https://infrahub-operator.github.io/vidra/api-references/) for detailed information on how to use Vidra effectively.

- [Official Documentation](https://infrahub-operator.github.io/vidra/)
- [Codespaces Demo](https://github.com/infrahub-operator/infrahub-vidra-demo)


## Contributing

We welcome contributions! If you have ideas, issues, or feature requests, please open an issue or submit a pull request on GitHub.

---

Thank you for using Vidra! 🎉
