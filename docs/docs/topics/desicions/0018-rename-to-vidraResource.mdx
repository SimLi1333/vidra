---
title: Refactoring Manifest Download, Resource Renaming, and Separation of Concerns in VidraResource
sidebar_position: 19
---
import Admonition from '@theme/Admonition';

# Refactoring Manifest Download, Resource Renaming, and Separation of Concerns in VidraResource

## Context and Problem Statement

After implementing the fourth GitOps principle, we realized that downloading the manifest could be handled within the `InfrahubSync` controller (formerly it was handled in the second controller called `InfrahubResource`). By saving the checksum of the artifact in the status section and checking it during each reconciliation, we ensure that the manifest is only downloaded when the checksum changes. (This logic was implemented once the second controller needed to run in a interval to ensure the actual system state matches the desired state in Infrahub.)

This change separates the concerns of connecting to Infrahub from managing resources on the Kubernetes cluster based on the manifest. And simplifies the `VidraResource` CRD by removing the manifest field from the status section and eliminating the need for dublicating a separate source section. Instead, the manifest can be stored in the `spec` of the `VidraResource`, aligning more with kubernetes best practices of storing desired state in the `spec`.

Additionally, this refactoring led to renaming the `InfrahubResource` to `VidraResource`, reflecting the broader scope and purpose of the resource and avoiding confution with the `InfrahubSync` CRD. The `VidraResource` now represents any Kubernetes resource that can be managed by Vidra, not just those related to Infrahub.

This refactoring also enables the creation of additional sync controllers for any other external systems in the future, increasing extensibility.

<Admonition type="note" title="Note">
This was a not considered during planning. We initially implemented the manifest download logic in the `InfrahubResource` controller, but as we progressed, we recognized the need for a cleaner separation of concerns and better alignment with Kubernetes best practices.
</Admonition>

## Considered Options

* **Keep manifest download in the main controller**
  One considered option is to keep the manifest download logic within the main controller. This approach would mix concerns by combining Infrahub-specific logic with Kubernetes resource management in the `Resource` controller. The manifest field would remain in the status section, and a separate source section would still be required in the CRD. While this option would not require additional implementation effort—since the current setup is already in place and tested—it would limit future extensibility. The controller would remain tightly coupled to Infrahub, making it harder to adapt for other external systems or expand the operator in the future.

* **Move manifest download to InfrahubSync**
  This option involves moving the manifest download logic to the `InfrahubSync` controller. The controller would handle downloading the manifest based on the same logic by storing the checksum in the status section, ensuring that the manifest is only downloaded when necessary. The `VidraResource` CRD would be simplified by removing the manifest field from the status section and eliminating the need for a source section in the `spec`. The `VidraResource` would then represent any Kubernetes resource managed by Vidra, not just those related to Infrahub.

## Decision Outcome

**Chosen option: Move manifest download to InfrahubSync, refactor CRD, and rename resource.**  
This approach cleanly separates concerns, simplifies the CRD, aligns with GitOps principles, and sets the foundation for supporting additional external systems.

### Consequences

* Good, because it improves maintainability, clarity, and separation of concerns, and allows for future extensibility.
* Good, because it aligns with Kubernetes best practices by storing the desired state in the `spec` and removing unnecessary duplication.
* Good, because it enables the creation of additional sync controllers for other external systems in the future, increasing extensibility.
* Bad, because it requires changes to the CRD, controller logic and tests, includes renaming a CRD in the Operator SDK, which is not intended by the SDK.