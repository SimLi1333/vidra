---
title: Programming Language
sidebar_position: 6
---

# Programming Language

## Context and Problem Statement

We need to choose programming languages for both the operator and the Infrahub components. The options are based on the supported languages of the Operator-SDK and the requirements of Infrahub.

## Considered Options

* **Go** 
    Best for the operator, as it is the primary language supported by the Operator-SDK. Go is well-suited for building Kubernetes operators due to its performance, concurrency model, and strong type system.

* **Python**
    Best for Infrahub, as it is widely used for application logic, scripting, and integrations by infrahub itselfe. 

## Decision Outcome

**Chosen options: Both Go and Python**, because:  
- **Go** for the operator, because go is used by the Operator-SDK, making it easier to implement complex logic and testing. Additionally, it integrates best with Kubernetes, allowing us to reuse Kubernetes native functions.
- **Python** for Infrahub, as Infrahub is primarily written in Python, which allows us to leverage existing libraries and tools. Python's flexibility and ease of use make it ideal for data transormation logic and scripting tasks within Infrahub.

This combination leverages the strengths of both languages for their respective domains.

### Consequences
* Good, because it allows us to use the best tools for each task, leveraging Go's performance and Kubernetes integration for the operator, while using Python's flexibility and ecosystem for Infrahub.
* Bad, because it requires team members to be proficient in both languages, which may lead to a steeper learning curve for some. However, we can mitigate this by providing training and resources for team members to learn Go.