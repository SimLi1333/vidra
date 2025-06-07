---
title: Documenting Vidra with Docusaurus
sidebar_position: 21
---
import Admonition from '@theme/Admonition';

# Documenting Vidra with Docusaurus

## Context and Problem Statement

We needed to choose a documentation framework for Vidra. Consistency of documentation across projects is important for maintainability and user experience.

## Considered Options

* **Docusaurus**  
    A popular documentation framework based on React, widely used in the Infrahub ecosystem. It allows us to maintain a consistent look and feel with other Infrahub tools.

* **MkDocs**  
    A simpler, static site generator that is easy to set up and use. It is lightweight and has a straightforward configuration, but it does not provide the same level of customization and integration as Docusaurus.

<Admonition type="note" title="Note">
We also used **crd-ref-docs** to generate API documentation for our Custom Resource Definitions (CRDs). This is integrated into our CI pipeline and published automatically to GitHub Pages but this is indipendent of this desicion.
</Admonition>

## Decision Outcome

**Chosen option: "Docusaurus"**, because it allows us to blend in with the documentation style of other Infrahub tools, ensuring a unified experience for users and contributors. While MkDocs was considered for its simplicity, consistency, integration, and style were prioritized.

### Consequences

* Good: Consistent documentation experience across all Infrahub projects, easier maintenance, and streamlined publishing via GitHub Actions.
* Bad: Slightly steeper learning curve compared to MkDocs, but mitigated by referencing existing usage of Docusaurus in Infrahub.

<Admonition type="note" title="Note">
For more information on how to use Docusaurus, refer to the [Docusaurus documentation](https://docusaurus.io/docs).
</Admonition>
<Admonition type="note" title="Note">
A search feature is added provided by <code>@easyops-cn/docusaurus-search-local</code> for local full-text search within the documentation.
</Admonition>
<Admonition type="note" title="Note">
A Blog feature is added for feature announcements and updates like the creation of a Demo environment.
</Admonition>