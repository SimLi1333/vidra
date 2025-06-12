---
title: Event-Driven vs. Owns-Based Reconciliation for Managed Resources
sidebar_position: 18
---
import Admonition from '@theme/Admonition';

# Event-Driven vs. Owns-Based Reconciliation for Managed Resources

## Context and Problem Statement

We sought to improve the responsiveness and efficiency of the Vidra Operator's reconciliation process. The traditional static requeue interval had to be quite tight; otherwise, it was not interactive enough, especially when users changed or deleted managed resources. We evaluated two main approaches to achieve a more event-driven mechanism, in line with GitOps principles.

<Admonition type="note" title="Note">
This was not considered during planning. We implemented the static requeue interval approach first, as the timeline was tight and we wanted to ensure the operator was functional. However, we recognized the need for a more dynamic and responsive solution that could handle changes in managed resources more effectively.
</Admonition>

## Considered Options

* **Owns-based scheduling (static, using `Owns()` in setup)**  
    By referencing all relevant GroupVersionKinds (GVKs) statically in the controller's setup with `Owns()`, the operator can watch for changes to managed resources. Filtering is required to avoid reconciling the operator's own updates of the resources. We could use predicates (e.g., generation change) and label selectors to target only relevant resource updates. This approach is straightforward but requires maintaining an up-to-date list of all GVKs and cannot be extended with new GVKs during operation. 
    
    Additionally, if a user edits a resource that already has the Vidra labels present, such updates are also filtered out by this approach. This means legitimate user-driven changes may not always trigger reconciliation, making the solution imperfect.

* **Event-driven reconciliation using [Informers](https://pkg.go.dev/k8s.io/client-go/informers) and GroupVersionResources (GVRs) in the controller setup function (semi-dynamic)**  
    This approach uses Informers to watch for changes to specific GroupVersionResources (GVRs) in the cluster. The Informers can be configured to trigger reconciliation when relevant events occur, such as creation, update, or deletion of resources. Calling it from the setup function allows the operator to ask Kubernetes for all GVRs with a specific label (e.g., `managed-by: vidra`). This allows the operator to not miss any changes on managed resources, or sub-resources, even if they are not explicitly defined in the manifest. It also ensures the Informers are not registered twice for the same resource, as the setup only runs one time. But a restart of the operator is required to add new GVRs.
    This approach could potentially be resource-intensive, as it requires maintaining Informers for all resources in the cluster with a label and may lead to performance issues if many resources are being watched. 

* **Event-driven reconciliation using [Informers](https://pkg.go.dev/k8s.io/client-go/informers) and GroupVersionResources (GVRs) called within the reconciliation function (dynamic)**  
    This approach dynamically creates Informers (Kubernetes-native event-based subscriptions to cluster events) for each managed resource's GVR. The Informers can be configured to watch for changes to only the specific resources which are reconciled and trigger reconciliation again when relevant events occur. This allows for more dynamic handling of Informers, as new GVRs can be added without needing to restart the operator.
    Problematically, this approach requires careful management of concurrency and thread safety, as multiple Informers could be added for the same GVR simultaneously, causing reconciliation loops. Mutexes and a map of watched GVRs are needed to ensure thread safety and avoid reconciliation loops. Additionally, generation predicates and labels can be used to filter relevant events and resources.

## Decision Outcome

**Chosen option: "Event-driven reconciliation using [Informers](https://pkg.go.dev/k8s.io/client-go/informers) and GroupVersionResources (GVRs) out of the reconciliation function (dynamic)"**, because it provides a more flexible and responsive reconciliation mechanism. This approach allows the operator to react to changes in managed resources immediately, without relying on static GVKs.
While being the most complex solution, it also allows for dynamic addition of new GVRs without requiring a restart of the operator, making it more adaptable to changes in the cluster.  
We implemented dynamic Informers with callback trigger functions (updating the corresponding `VidraResource`) for each change in a managed resource. Thread safety is ensured with mutexes and a map of watched GVRs. Generation predicates and labels are used to filter relevant events and resources. The factory pattern and function callbacks provide flexibility for handling triggers, maintaining a clean separation of concerns and allowing for easy extension in the future.

### Consequences

* Good, because it is truly event-driven, allowing for immediate reaction to changes in managed resources, improving responsiveness and user experience.
* Good, because it allows for dynamic addition of new GVRs without requiring a restart of the operator, making it more flexible and adaptable to changes in the cluster.
* Good, because it aligns with GitOps principles by ensuring that the operator only reconciles resources that are relevant and managed by Vidra, avoiding unnecessary overhead.
* Bad, because it increases complexity, requires careful concurrency management, and makes it harder to avoid reconciliation loops. 
