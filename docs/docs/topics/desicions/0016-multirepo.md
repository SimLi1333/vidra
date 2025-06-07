---
title: Monorepo vs Multirepo for Vidra, CLI, and Documentation
sidebar_position: 17
---

# Monorepo vs Multirepo for Vidra, CLI, and Documentation

## Context and Problem Statement

While moving the code from Ost gitlab to gitHub, we needed to decide whether to maintain separate repositories for Vidra, the CLI, and the documentation, or to combine them into a single repository.

While multiple repositories can provide isolation, they introduce complexity in versioning and coordination. Our use case requires tight coupling between the CLI and Vidra, and consistent documentation updates alongside releases.

## Considered Options

* **Multiple repositories**  
    Each component (Vidra, CLI, documentation) in its own repository. This increases isolation but complicates versioning, dependency management, and release coordination.

* **Single monorepo**  
    All components in one repository. This simplifies versioning, dependency management, and ensures consistency across all parts of the project.

## Decision Outcome

**Chosen option: "single monorepo"**, because the CLI is a supporting tool tightly coupled with Vidra's version, and a monorepo allows us to release all components together in a single pipeline under one coherent version.

### Consequences

* Good, because all development, releases, and documentation updates are managed in one place, making coordination and integration easier.
* Bad, because the repository may grow larger and require clear structure and ownership to avoid confusion.