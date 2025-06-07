---
title: Continuos Integration
sidebar_position: 7
---
import Admonition from '@theme/Admonition';

# Continuous Integration

We use **GitHub Actions** as our CI tool to automate testing, building, releasing and documentation deployment for the **Vidra Operator**. Currently, we maintain three workflows:

<div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
  <div>
    <ul>
      <li><strong>Deploy Docs</strong>(Deploy Docusaurus to GitHub Pages)</li>
      <li><strong>Operator CI</strong></li>
      <li><strong>Operator CLI</strong> (CLI Tool Build and Sign)</li>
      <li><strong>Code Quality and Security Checks</strong> (CodeQL, Depandabot Updates)</li>
    </ul>
  </div>
  <img src={require('../../static/img/workflow.png').default} alt="Workflows" style={{ height: 250, marginRight: 150 }} />
</div>

## Deploy Docs

The **Deploy Docs** workflow automates the deployment of our Docusaurus documentation site to GitHub Pages whenever documentation files are updated. In addition to publishing the static site, it generates up-to-date API documentation for our Custom Resource Definitions (CRDs) using `crd-ref-docs` (invoked via a `Make` target). The workflow builds the static site—including the auto-generated docs and site—and pushes the result to the `gh-pages` branch, ensuring that users always have access to the latest documentation. We also had to ensure the published helm release files are not overwritten, as they too live in the `gh-pages`.

### Trigger

- Any push to the `main` branch that affects documentation files.
- Specifically triggered when:
  - A file in the `docs/` directory is changed.

Check the [`docs.yaml`](https://github.com/infrahub-operator/vidra/blob/main/.github/workflows/docs.yaml) workflow file for detailed configuration.

---

## Operator CI

This workflow handles linting, testing, building, publishing and releasing the Vidra Operator to the GitHub Container Registry, GitHub Releases and GitHub Pages.

### Trigger

- Every push to a pull request or push to the `main` branch (protected).
- If any file coresponding to the operator is changed.
- Pushing a version tag (e.g., `v0.0.3`) will trigger the build and release process. Version is also tracked in the `Makefile`.

![CI Pipeline Overview](../../static/img/ci.png)

### Lint and Test

- Uses **golangci-lint** to perform static code analysis and linting on the Vidra Operator codebase.
- Executes all tests for the Vidra Operator using `go test` and Ginkgo-based test suites.
- Tests are run using `envtest` to provide a realistic Kubernetes API server.
- A test coverage report is generated and uploaded to **Codecov**.
- Pull requests where the code coverage drops below **80%** or the new line of codes are not covered by tests will have a warning comment added by the `codecov` GitHub Agent.

  <Admonition type="note" title="Note">
  The Codecov badge in the README reflects the current coverage status.
  </Admonition>

### Build and Publish the Operator

- Uses the official **Docker GitHub Actions** to build the Vidra Operator image.
- Pushes the image to the **GitHub Container Registry (GHCR)**.
- The image is tagged with the current version (e.g., `v0.0.3`).
- The operator bundle is generated using `Make bundle-build` and published to the GitHub Container Registry (for OLM install).
- The images are signed using **cosign**, allowing users to verify its authenticity and integrity.

### Helm Chart
- Helmify is used to generate a Helm chart and a build a Helm Release from the operator manifests and kustomize files. The Helm Release is also published to the GitHub Container Registry. (OCI install)
- The Helm chart is versioned and tagged with the same version as the operator image.
- The Helm chart Release (index.yaml and Release) is automatically published to the GitHub Releases, making it easy for users to install different versions of the Vidra Operator using Helm add repository commands.
- The packaged Helm Release and its index.yaml are also published to the `gh-pages` branch of the repository for easy access via one consistent Helm repository URL.

### Add Yaml files to the release
- The yaml files used for the install using OLM (Operator Lifecycle Manager) are added to the release with the correct version as well. This allows users to install the operator using OLM by applying the `catalogsource.yaml` and `subscription.yaml` files. 


You can find the workflow definition in [`main.yaml`](https://github.com/infrahub-operator/vidra/blob/main/.github/workflows/main.yaml).

---
## Operator CLI
This workflow is responsible for building and testing the Vidra Operator CLI tool, which provides a command-line interface for interacting with the Vidra Operator.

### Trigger
- Every push to a pull request or push to the `main` branch (protected).
- If any file corresponding to the operator CLI is changed.

### Lint and Test
- Uses **golangci-lint** to perform static code analysis and linting on the Vidra Operator CLI codebase.
- Executes all tests for the Vidra Operator CLI using `go test` and Ginkgo-based test suites.

### Build and Publish the CLI
- Uses the official **Docker GitHub Actions** to build the Vidra Operator CLI image.
- Pushes the image to the **GitHub Container Registry (GHCR)**.
- The image is tagged with the current version (e.g., `v0.0.3`).
- The CLI is built and packaged as a binary, which is also published to the GitHub Releases.
- The CLI binary is signed using **cosign**, allowing users to verify its authenticity and integrity.

You can find the workflow definition in [`cli.yaml`](https://github.com/infrahub-operator/vidra/blob/main/.github/workflows/cli.yaml).
---

## Code Quality and security Checks

- **Dependabot** is configured to automatically create pull requests for dependency updates of packages used in the repository.
- **CodeQL** is used to perform static code analysis and security checks on the codebase. It scans for vulnerabilities and potential issues in the code, helping maintain a high standard of code quality and security.
---

