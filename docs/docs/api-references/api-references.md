# API Reference

## Packages
- [infrahub.operators.com/v1alpha1](#infrahuboperatorscomv1alpha1)


## infrahub.operators.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the infrahub v1alpha1 API group

### Resource Types
- [InfrahubSync](#infrahubsync)
- [VidraResource](#vidraresource)



#### InfrahubSync



InfrahubSync is the Schema for the infrahubsyncs API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `infrahub.operators.com/v1alpha1` | | |
| `kind` _string_ | `InfrahubSync` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[InfrahubSyncSpec](#infrahubsyncspec)_ | Spec defines the desired state of InfrahubSync |  |  |
| `status` _[InfrahubSyncStatus](#infrahubsyncstatus)_ | Status defines the observed state of InfrahubSync |  |  |


#### InfrahubSyncDestination



VidraResourceDestination contains information about where the resource will be sent



_Appears in:_
- [InfrahubSyncSpec](#infrahubsyncspec)
- [VidraResourceSpec](#vidraresourcespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `server` _string_ | Only needed if you need to deploy to two Kubernetis cluster (multicluster) if set to "httlps://kubernetes.default.svc" or omitted, the operator will use the current cluster |  | Optional: \{\} <br /> |
| `namespace` _string_ | Default Namespace in the Kubernetes cluster where the resource should be sent, if they do not hava a namespace already set |  | Optional: \{\} <br /> |
| `reconcileOnEvents` _boolean_ | If true, the operator will reconcile resources based on k8s events. (default: false) - changes to the resource will trigger a reconciliation | false |  |


#### InfrahubSyncSource



VidraResourceSource contains the source information for the resource



_Appears in:_
- [InfrahubSyncSpec](#infrahubsyncspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `infrahubAPIURL` _string_ | URL for the Infrahub API (e.g., https://infrahub.example.com) |  | Pattern: `^(http\|https)://[a-zA-Z0-9.-]+(:[0-9]+)?(?:/[a-zA-Z0-9-]+)*$` <br />Required: \{\} <br /> |
| `targetBranch` _string_ | The target branch in Infrahub to interact with |  | MinLength: 1 <br />Required: \{\} <br /> |
| `targetDate` _string_ | The target date in Infrahub for all the interactions (e.g., "2025-01-01T00:00:00Z or -2d" for the artifact from two days ago) |  | Format: date-time <br />Optional: \{\} <br /> |
| `artefactName` _string_ | Artifact name that is being handled by the operator, this is used to identify the resource in Infrahub |  | MinLength: 1 <br />Required: \{\} <br /> |


#### InfrahubSyncSpec



InfrahubSyncSpec defines the desired state of InfrahubSync



_Appears in:_
- [InfrahubSync](#infrahubsync)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `source` _[InfrahubSyncSource](#infrahubsyncsource)_ | Foo is an example field of InfrahubSync. Edit infrahubsync_types.go to remove/update<br />Source contains the source information for the Infrahub API interaction |  |  |
| `destination` _[InfrahubSyncDestination](#infrahubsyncdestination)_ | Destination contains the destination information for the resource |  |  |


#### InfrahubSyncStatus



InfrahubSyncStatus defines the observed state of InfrahubSync



_Appears in:_
- [InfrahubSync](#infrahubsync)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `checksums` _string array_ | Checksums contains a list of checksums for synced resources |  | Enum: [Pending Running Succeeded Failed Stale] <br /> |
| `syncState` _[State](#state)_ | SyncState indicates the current state of the sync operation |  |  |
| `lastError` _string_ | LastError provides details about the last error encountered during the sync operation |  |  |
| `lastSyncTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta)_ | LastSyncTime indicates the last time the sync operation was performed |  |  |


#### ManagedResourceStatus







_Appears in:_
- [VidraResourceStatus](#vidraresourcestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind of the resource (e.g., Deployment, Service) |  |  |
| `apiVersion` _string_ | APIVersion of the resource (e.g., apps/v1) |  |  |
| `name` _string_ | Name of the resource |  |  |
| `namespace` _string_ | Namespace of the resource |  |  |


#### State

_Underlying type:_ _string_





_Appears in:_
- [InfrahubSyncStatus](#infrahubsyncstatus)
- [VidraResourceStatus](#vidraresourcestatus)

| Field | Description |
| --- | --- |
| `Running` |  |
| `Succeeded` |  |
| `Failed` |  |
| `Stale` |  |


#### VidraResource



VidraResource is the Schema for the Vidraresources API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `infrahub.operators.com/v1alpha1` | | |
| `kind` _string_ | `VidraResource` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[VidraResourceSpec](#vidraresourcespec)_ | VidraResourceSpec defines the desired state of VidraResource |  |  |
| `status` _[VidraResourceStatus](#vidraresourcestatus)_ | VidraResourceStatus defines the observed state of VidraResource |  |  |


#### VidraResourceSpec



VidraResourceSpec defines the desired state of VidraResource



_Appears in:_
- [VidraResource](#vidraresource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `destination` _[InfrahubSyncDestination](#infrahubsyncdestination)_ | Destination contains the destination information for the resource |  |  |
| `manifest` _string_ | Manifest contains the manifest information for the resource |  |  |
| `reconcileOnEvents` _boolean_ | If true, the operator will reconcile resources based on k8s events. (default: false) | false |  |
| `reconciledAt` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta)_ | The last time the resource was reconciled |  |  |


#### VidraResourceStatus



VidraResourceStatus defines the observed state of VidraResource



_Appears in:_
- [VidraResource](#vidraresource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `managedResources` _[ManagedResourceStatus](#managedresourcestatus) array_ | ManagedResources contains a list of resources managed by this VidraResource |  |  |
| `DeployState` _[State](#state)_ | DeployState indicates the current state of the deployment |  | Enum: [Pending Running Succeeded Failed Stale] <br /> |
| `lastError` _string_ | LastError contains the last error message if any |  |  |
| `lastSyncTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta)_ | LastSyncTime indicates the last time the resource was synchronized |  |  |


