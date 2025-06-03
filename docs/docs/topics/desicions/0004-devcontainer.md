---
title: Dev Container
sidebar_position: 5
---


# Dev Container

## Context and Problem Statement

As in every projet ther is the decision whether to use a dev container for environment consistency or rely on manual setup, considering the team's operating systems and the simplicity of the required tooling.

## Considered Options

* **Dev Container**
  Offers a consistent environment for all developers but adds setup complexity.

* **Manual Setup**
    Simple and straightforward, but may lead to inconsistencies across team members' environments.

## Decision Outcome

**Chosen option: "Manual Setup"**, because setting up the environment only requires installing Operator SDK, and all team members use Linux or Mac, which minimizes compatibility issues. While dev containers offer consistency and isolation, the simplicity of our requirements and the team's homogeneous environment make manual setup more practical.

### Consequences
* Good, because it allows for a quick start without the overhead of managing a dev container.
* Bad, because it may lead to inconsistencies if new contributors use different setups or operating systems in the future. However, this risk is mitigated by the simplicity the installation of the required tools using Makefile targets.