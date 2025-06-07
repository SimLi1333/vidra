---
title: Multi-Cluster Feature for Vidra Operator
sidebar_position: 12
---
import Admonition from '@theme/Admonition';

# Multi-Cluster Feature for Vidra Operator

## Context and Problem Statement

We wanted to enable Vidra Operator to manage resources across multiple Kubernetes clusters from a single installation. Previously, the operator had to be installed on every cluster to reconcile resources, which increased operational overhead.


<Admonition type="note" title="Note">
This was a nice-to-have feature, not a must-have during planing. If there was no time to implement it, we would have left it out.
</Admonition>

## Considered Options

* **Install operator on every cluster**  
    Each cluster runs its own instance of Vidra Operator, managing only local resources.

* **Single operator managing multiple clusters**  
    One Vidra Operator instance manages resources on multiple clusters by connecting to the kubernetis API remotely.

## Decision Outcome

**Chosen option: "Single operator managing multiple clusters"**, because it simplifies management by centralizing the operator in one place, reducing the need for multiple installations and configurations. This approach allows for easier updates, consistent behavior across clusters, and a unified view of all managed resources.

We implemented a mechanism where Vidra Operator can deploy managed resources on remote clusters. This is achieved by providing a new Secret with the destination server as label, containing the kubeconfig for the target cluster. On the first reconciliation of a resource destined for a different cluster, a dynamic Go client is created using the kubeconfig and destination server. To optimize performance and resource usage, this client is cached and reused for subsequent reconciliations, using mutexes to ensure thread safety.

This design doesn not exclude the possibility of installing the operator on each cluster if desired. It provides flexibility for users who prefer to manage resources locally while still allowing centralized management.

### Consequences

* Good, because it reduces the need to install and maintain the operator on every cluster, centralizes management, and offers flexibility.  
* Bad, because it introduces complexity in client management, requires secure handling of kubeconfigs, and needs careful access control to avoid security risks.  

