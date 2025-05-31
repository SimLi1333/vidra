---
title: Event-Driven vs. Owns-Based Reconciliation for Managed Resources
sidebar_position: 18
---
import Admonition from '@theme/Admonition';

# Event-Driven vs. Owns-Based Reconciliation for Managed Resources

## Context and Problem Statement

We sought to improve the responsiveness and efficiency of the Vidra Operator's reconciliation process. The traditional static requeue interval had to be quite tight otherwise it was not interactive enough, especially when users change or delete managed resources. We evaluated two main approaches to achieve a more event-driven mechanism, in line with GitOps principles.

<Admonition type="note" title="Note">
This was a not considered during planning. We almost implemented the static requeue interval approach, as the timeline was tight and we wanted to ensure the operator was functional. However, we recognized the need for a more dynamic and responsive solution that could handle changes in managed resources more effectively.
</Admonition>

## Considered Options

* **Owns-based scheduling (static, using `Owns()` in setup)**  
    By referencing all relevant GroupVersionKinds (GVKs) statically in the controller's setup with `Owns()`, the operator can watch for changes to managed resources. Filtering is required to avoid reconciling of the operator's own updates ofthe resources. We could use a predicates (e.g., generation change) and label selectors to target only relevant resources. This approach is straightforward but requires maintaining an up-to-date list of all GVKs and can not be extended with new GVKs during operation. 
    
    Additionally, if a user edits a resource that already has the Vidra labels present, such updates are also filtered out by this approach. This means legitimate user-driven changes may not always trigger reconciliation, making the solution imperfect.

* **Event-driven reconciliation using Informers and GroupVersionResources (GVRs) in the setup (semi dinamic)**  
    This approach uses Informers to watch for changes to specific GroupVersionResources (GVRs) in the cluster. The Informers can be configured to trigger reconciliation when relevant events occur, such as creation, update, or deletion of resources. Calling it from the setup function allows the operator to ask Kubernetes for all GVRs with a specific label (e.g., `managed-by: vidra`). This allows the operator to not miss any changes on manageed resources, or sub-resources, even if they are not explicitly defined in the manifest. But a restart of the operator is required to add new GVRs.
    Thia approach could potentially be resource-intensive, as it requires maintaining Informers for all managed resources and may lead to performance issues if many resources are being watched. 

* **Event-driven reconciliation using Informers and GroupVersionResources (GVRs) called within the reconciliation function (dynamic)**  
    This approach dynamically creates Informers (kubernetes native event-based subsciption to cluster events) for each managed resource's GVR. The Informers can be configured to watch for changes to specific resources and trigger reconciliation when relevant events occur. This allows for more dynamic handling of resources, as new GVRs can be added without needing to restart the operator.
    Problematically, this approach requires careful management of concurrency and thread safety, as multiple Informers may be running simultaneously. Mutexes and a registry of watched GVRs are needed to ensure thread safety and avoid recanciliation loops. Additionally, generation predicates and labels can be used to filter relevant events and resources.


## Decision Outcome

**Chosen option: "Event-driven reconciliation using Informers and GroupVersionResources (GVRs) out of the reconciliation function (dynamic)"**, because it provides a more flexible and responsive reconciliation mechanism. This approach allows the operator to react to changes in managed resources immediately, without relying on static requeue intervals or filtering out updates that should not trigger reconciliation.  
We implemented dynamic Informers with callback trigger functions (updating the according `VidraResource`) for each change in the managed resource. Thread safety is ensured with mutexes and a registry of watched GVRs. Generation predicates and labels are used to filter relevant events and resources. The factory pattern and function callbacks provide flexibility for handling triggers, maintaining a clean separation of concerns and allowing for easy extension in the future.

### Consequences

* Good, because it is truly event-driven, allowing for immediate reaction to changes in managed resources, improving responsiveness and user experience.
* Good, because it allows for dynamic addition of new GVRs without requiring a restart of the operator, making it more flexible and adaptable to changes in the cluster.
* Good, because it aligns with GitOps principles by ensuring that the operator only reconciles resources that are relevant and managed by Vidra, avoiding unnecessary overhead.
* Bad, because it increases complexity, requires careful concurrency management, and makes it harder to avoid reconciliation loops. 
