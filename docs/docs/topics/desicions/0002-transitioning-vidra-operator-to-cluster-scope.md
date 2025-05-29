# Transitioning Vidra Operator to Cluster Scope

## Context and Problem Statement

We needed to decide whether the Controller should be namespace-scoped or cluster-scoped.

While namespace-scoped Operators offer isolation, they are limited as they can not take ownership of resources in other namespaces. Our use case requires managing resources across multiple namespaces and potentially work with ownership could improve this.

## Considered Options

* Namespace-scoped controller and CRDs
* cluster-scoped controller and CRD

## Decision Outcome

Chosen option: "cluster-scoped controller and CRD", because While namespace-scoped is possible especaly in combination with finalizers to delete resources again, we do not want to lose the benefits of kubernetes ownership. We chose to go with cluster-scoped and leave us all options open to use ownership or finalizers to remove managed resources again.
