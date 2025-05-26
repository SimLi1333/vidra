# API Reference

## Packages
- [vidra.operators.com/v1alpha1](#vidraoperatorscomv1alpha1)


## vidra.operators.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the infrahub v1alpha1 API group

### Resource Types
- [InfrahubResource](#infrahubresource)
- [InfrahubSync](#infrahubsync)



#### InfrahubResource



InfrahubResource is the Schema for the infrahubresources API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `vidra.operators.com/v1alpha1` | | |
| `kind` _string_ | `InfrahubResource` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[InfrahubResourceSpec](#infrahubresourcespec)_ | InfrahubResourceSpec defines the desired state of InfrahubResource |  |  |
| `status` _[InfrahubResourceStatus](#infrahubresourcestatus)_ | InfrahubResourceStatus defines the observed state of InfrahubResource |  |  |


#### InfrahubResourceIDs



InfrahubResourceIDs contains identifiers for the resource



_Appears in:_
- [InfrahubResourceSpec](#infrahubresourcespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `artefactID` _string_ | Unique identifier for the artifact |  | Required: \{\} <br /> |
| `checksum` _string_ | Checksum of the artifact |  | Required: \{\} <br /> |
| `storageID` _string_ | Storage ID for the artifact |  | Required: \{\} <br /> |


#### InfrahubResourceSpec



InfrahubResourceSpec defines the desired state of InfrahubResource



_Appears in:_
- [InfrahubResource](#infrahubresource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `source` _[InfrahubSyncSource](#infrahubsyncsource)_ | Source contains the source information for the Infrahub API interaction |  |  |
| `destination` _[InfrahubSyncDestination](#infrahubsyncdestination)_ | Destination contains the destination information for the resource |  |  |
| `ids` _[InfrahubResourceIDs](#infrahubresourceids)_ | IDs contains important identifiers for the resource |  |  |


#### InfrahubResourceStatus



InfrahubResourceStatus defines the observed state of InfrahubResource



_Appears in:_
- [InfrahubResource](#infrahubresource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `managedResources` _[ManagedResourceStatus](#managedresourcestatus) array_ | ManagedResources contains the status of managed resources |  |  |
| `DeployState` _[State](#state)_ | DeployState indicates the current state of the deployment |  | Enum: [Pending Running Succeeded Failed Stale] <br /> |
| `lastError` _string_ | LastError contains the last error message if any |  |  |
| `lastSyncTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta)_ | LastSyncTime indicates the last time the resource was synchronized |  |  |


#### InfrahubSync



InfrahubSync is the Schema for the infrahubsyncs API





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `vidra.operators.com/v1alpha1` | | |
| `kind` _string_ | `InfrahubSync` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[InfrahubSyncSpec](#infrahubsyncspec)_ | Spec defines the desired state of InfrahubSync |  |  |
| `status` _[InfrahubSyncStatus](#infrahubsyncstatus)_ | Status defines the observed state of InfrahubSync |  |  |


#### InfrahubSyncDestination



InfrahubResourceDestination contains information about where the resource will be sent



_Appears in:_
- [InfrahubResourceSpec](#infrahubresourcespec)
- [InfrahubSyncSpec](#infrahubsyncspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `server` _string_ | Server URL for the destination (usually a Kubernetes API endpoint) |  | Optional: \{\} <br /> |
| `namespace` _string_ | Namespace in the Kubernetes cluster where the resource should be sent |  | Optional: \{\} <br /> |


#### InfrahubSyncSource



InfrahubResourceSource contains the source information for the resource



_Appears in:_
- [InfrahubResourceSpec](#infrahubresourcespec)
- [InfrahubSyncSpec](#infrahubsyncspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `infrahubAPIURL` _string_ | URL for the Infrahub API |  | Pattern: `^(http\|https)://[a-zA-Z0-9.-]+(:[0-9]+)?(?:/[a-zA-Z0-9-]+)*$` <br />Required: \{\} <br /> |
| `targetBranch` _string_ | The target branch to interact with |  | MinLength: 1 <br />Required: \{\} <br /> |
| `targetDate` _string_ | The target date for the request |  | Format: date-time <br />Optional: \{\} <br /> |
| `artefactName` _string_ | Artifact name that is being handled |  | MinLength: 1 <br />Required: \{\} <br /> |


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
| `syncState` _[State](#state)_ | SyncState indicates the current state of the synchronization process |  | Enum: [Pending Running Succeeded Failed Stale] <br /> |
| `lastError` _string_ | LastError contains the last error message if any |  |  |
| `lastSyncTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.31/#time-v1-meta)_ | LastSyncTime indicates the last time the synchronization was performed |  |  |


#### ManagedResourceStatus







_Appears in:_
- [InfrahubResourceStatus](#infrahubresourcestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind of the resource (e.g., Deployment, Service) |  |  |
| `apiVersion` _string_ | APIVersion of the resource (e.g., apps/v1) |  |  |
| `name` _string_ | Name of the resource |  |  |
| `namespace` _string_ | Namespace of the resource |  |  |


#### State

_Underlying type:_ _string_





_Appears in:_
- [InfrahubResourceStatus](#infrahubresourcestatus)
- [InfrahubSyncStatus](#infrahubsyncstatus)

| Field | Description |
| --- | --- |
| `Running` |  |
| `Succeeded` |  |
| `Failed` |  |
| `Stale` |  |


