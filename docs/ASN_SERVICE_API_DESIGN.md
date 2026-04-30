# ASN Service API Design

> Defines the full interface contract between the ASN framework and service implementations.  
> Covers: call sequencing, re-entrancy guarantees, state machines, and implementation obligations.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Package Organization](#2-package-organization)
3. [Object Relationships and Ownership](#3-object-relationships-and-ownership)
4. [State Management](#4-state-management)
5. [Controller Side — `capi`](#5-controller-side--capi)
   - 5.1 [ASNServiceController (service-implemented)](#51-asnservicecontroller-service-implemented)
   - 5.2 [ASNController (framework-provided)](#52-asncontroller-framework-provided)
6. [Service Node Side — `snapi`](#6-service-node-side--snapi)
   - 6.1 [ASNService (service-implemented)](#61-asnservice-service-implemented)
   - 6.2 [ASNServiceNode (framework-provided)](#62-asnservicenode-framework-provided)
7. [Config Ops](#7-config-ops)
8. [Ops Commands](#8-ops-commands)
9. [Cross-Service Data Sharing](#9-cross-service-data-sharing)
10. [IAM](#10-iam)
11. [Subscription](#11-subscription)
12. [Common Abstractions](#12-common-abstractions)
13. [Network and Node Topology](#13-network-and-node-topology)
14. [Implementation Checklist](#14-implementation-checklist)

---

## 1. Architecture Overview

ASN is a distributed control-plane framework. Its topology is fixed: **one Controller** (centralized management plane) paired with **N Service Nodes** (data plane), each running on a distinct host. Services are loaded as **shared libraries (`.so`)** at runtime.

```
┌────────────────────────────────────────────────────────┐
│                     ASN Framework                      │
│                                                        │
│   ┌──────────────────────────────┐  (single instance)  │
│   │        ASN Controller        │                     │
│   │  impl:  ASNController        │                     │
│   │  calls: ASNServiceController │                     │
│   └──────────────┬───────────────┘                     │
│                  │ manages N nodes                     │
│   ┌──────────────▼───────────────┐  (one per node)     │
│   │     ASN Service Node ×N      │                     │
│   │  impl:  ASNServiceNode       │                     │
│   │  calls: ASNService           │                     │
│   └──────────────────────────────┘                     │
└────────────────────────────────────────────────────────┘
```

The framework owns all topology (networks, nodes, groups). Services observe and annotate it. Services do not communicate through the framework's network layer; intra-node cross-service data exchange uses the Shared Data mechanism (§9).

---

## 2. Package Organization

| Package | Path Suffix | Role |
|---|---|---|
| `capi` | `/controller` | Controller-side interfaces and structs |
| `snapi` | `/servicenode` | Service-node-side interfaces and structs |
| `commonapi` | `/common` | Shared enums, structs, DB/log abstractions |
| `iam` | `/iam` | IAM interface |
| `subscription` | `/subscription` | Subscription / IAP interface |
| `log` | `/log` | Structured logger interface |

---

## 3. Object Relationships and Ownership

```
  SERVICE-IMPLEMENTED                   FRAMEWORK-PROVIDED
  ───────────────────                   ──────────────────

  ┌─────────────────────┐               ┌─────────────────────┐
  │ ASNServiceController│◄─ lifecycle ──│                     │
  │                     │              │   ASNController      │
  │                     │──── calls ──►│                     │
  └─────────────────────┘               └─────────────────────┘

  ┌─────────────────────┐               ┌─────────────────────┐
  │     ASNService      │◄─ lifecycle ──│                     │
  │                     │              │   ASNServiceNode     │
  │                     │──── calls ──►│                     │
  └─────────────────────┘               └─────────────────────┘
```

**Ownership rules:**

| Interface | Implemented by | Consumed by |
|---|---|---|
| `ASNController` | Framework | Your `ASNServiceController` |
| `ASNServiceController` | You | Framework |
| `ASNServiceNode` | Framework | Your `ASNService` |
| `ASNService` | You | Framework |
| `iam.Instance` | Framework | Your `ASNServiceController` (via `GetIAM()`) |
| `subscription.Instance` | Framework | Your `ASNServiceController` (via `GetSubscription()`) |

Both `ASNServiceController` and `ASNService` expose a `StaticResource` sub-interface that the framework may call **before** `Init()`.

---

## 4. State Management

### Node States (`commonapi.NodeState`)

Node state is framework-owned; services cannot set it directly.

| Value | Constant | Description |
|---|---|---|
| `0` | `NodeStateUnregistered` | Never successfully registered |
| `1` | `NodeStateOffline` | Registered but currently unreachable |
| `2` | `NodeStateOnline` | Connected and reachable |
| `3` | `NodeStateMaintenance` | Online but in maintenance mode |

```
Unregistered ──► Offline ◄──► Online ◄──► Maintenance
```

`NodeStateChange` events carry both the updated `NodeState` and the current `ServiceState` for that node. Events also include `FrameworkError` (non-nil on framework-level failures, e.g. `FrameworkErrNodeDisconnected` when a node goes offline or `FrameworkErrServiceTimeout` on timeout) and `ServiceError` (non-nil when the service itself reported an error during a state transition). When a node transitions to `Offline`, any in-flight ops targeting it yield `FrameworkErrNodeDisconnected`.

### Service States (`commonapi.ServiceState`)

The framework tracks service state **independently per node**.

| Value | Constant | Description |
|---|---|---|
| `0` | `ServiceStateUnavailable` | `.so` not loaded |
| `1` | `ServiceStateUninitialized` | Loaded; `Init()` not yet succeeded |
| `2` | `ServiceStateInitialized` | `Init()` succeeded |
| `3` | `ServiceStateConfiguring` | Applying config or config ops |
| `4` | `ServiceStateRunning` | `Start()` succeeded; fully operational |
| `5` | `ServiceStateMalfunctioning` | Fatal error or config op failure |

```
Unavailable
   │  AddServiceToNode
   ▼
Uninitialized
   │  Init() succeeds
   ▼
Initialized ◄─── StopService() succeeds (from Running / Configuring / Malfunctioning)
   │  Start() called
   ▼
Configuring ──► Running ──► Configuring   (config update cycle)
                   │
                   │  runtimeErrChan fires / config op error / Stop() error
                   ▼
             Malfunctioning
```

**State influence mechanisms (service side):**

| Mechanism | Effect |
|---|---|
| `Init()` returns error | Stays `Uninitialized` |
| `Start()` returns `ErrRestartNeeded` | Framework calls `Stop()` + `Start()` with new config (no re-initialization) |
| `Start()` returns any other error | → `Malfunctioning` |
| `runtimeErrChan` receives a value | → `Malfunctioning` |
| `AddConfigOps` / `UpdateConfigOp` / `DeleteConfigOps` return error | → `Malfunctioning` |

**`FrameworkError` values in `OpsResponse`:**

| Error | Cause |
|---|---|
| `FrameworkErrServiceTimeout` | Node reachable but service did not respond within timeout |
| `FrameworkErrNodeDisconnected` | Node offline |
| `FrameworkErrServiceUnavailable` | Service not loaded on target node |
| `FrameworkErrServiceStateNotAllowed` | Service not in `Running` state |

When `FrameworkError != nil`, `ServiceResponse` and `ServiceError` are undefined.

### Config Source (`commonapi.ServiceSource`)

`Node.ServiceInfo.ConfigSource` indicates the origin of a node's active configuration:

| Constant | Meaning |
|---|---|
| `ServiceConfigSourceNode` | Config is set directly on the node |
| `ServiceConfigSourceNodeGroup` | Config is inherited from the node's group |

---

## 5. Controller Side — `capi`

### 5.1 ASNServiceController (service-implemented)

#### Lifecycle Sequence

```
StaticResource()             any time, including pre-Init; re-entrant
  ├─ Version()
  ├─ CLICommands()
  └─ WebHandler()

Init(asnController)          once; not re-entrant
Start(config)                after Init; sequential; may be called multiple times

  ├─ HandleMessageFromNode() after Init; concurrent
  ├─ AddConfigOps()          after Init; concurrent
  ├─ UpdateConfigOp()        after Init; concurrent
  ├─ DeleteConfigOps()       after Init; concurrent
  └─ GetMetrics()            after Init; concurrent

Stop()                       idempotent
Finish()                     once; after Stop
```

#### Method Contracts

---

**`StaticResource() StaticResource`**

- Pre-`Init()` safe; re-entrant. Returns a stable, immutable object.

---

**`StaticResource.Version() commonapi.Version`**

- Returns a hardcoded version constant.

---

**`StaticResource.CLICommands(applyCLIOps func(ServiceScope, []string, string, string) error) []*cobra.Command`**

- Returns the Cobra commands to register in the ASN CLI. Purely declarative — no resource allocation, no goroutines.
- `applyCLIOps` is the framework-provided dispatcher; call it from command handlers to fan ops to service nodes. Its semantics are identical to `ASNController.SendServiceOps`.

---

**`StaticResource.WebHandler(staticPath string) (http.Handler, error)`**

- Returns the HTTP handler for this service. The framework mounts it under a service-specific path prefix.
- `staticPath` points to the service's static asset directory on the filesystem.
- No heavy initialization here; defer resource acquisition to `Init()`.

---

**`Init(asnController ASNController) error`**

- Called once after `.so` load, before `Start()`. Not re-entrant.
- All `Init*` / `Get*` calls on `asnController` must occur here — they are one-shot and return an error on subsequent calls.
- On success, CLI commands must be immediately runnable.
- Background goroutines must not start here; defer to `Start()`.
- Failure leaves the service in `ServiceStateUninitialized`.

---

**`Start(config string) error`**

- Called after `Init()` and on every subsequent config change. Sequential (never concurrent with itself), but may be called multiple times.
- Must return promptly; start background goroutines as needed.
- Each invocation must fully supersede the previous config. Tear down and re-create stateful components if necessary.

---

**`AddConfigOps(serviceScope ServiceScope, scopeID string, configParams []string) error`**  
**`UpdateConfigOp(serviceScope ServiceScope, scopeID, configOpID, configParam string) error`**  
**`DeleteConfigOps(serviceScope ServiceScope, scopeID string, configOpIDs []string) error`**

- Called after `Init()`, potentially concurrent. Guard shared state.
- The framework has already persisted the ops and dispatched them to affected service nodes before invoking these. Implement controller-side bookkeeping here (routing tables, in-memory policy, etc.).
- `serviceScope` is `ServiceScopeNodeGroup` (3) or `ServiceScopeNode` (4); `scopeID` is the corresponding ID.

> **Design note (under discussion):** Whether this callback is invoked is under discussion. An alternative approach — consistent with how `Start()` errors are handled — is to omit controller-side pre-validation entirely and let config op errors propagate directly to the node, which then transitions to `Malfunctioning`. The callbacks may be removed in a future revision.

---

**`HandleMessageFromNode(nodeID, messageType, payload string) error`**

- Handles upcalls from service nodes sent via `ASNServiceNode.SendMessageToController()`.
- Called after `Init()`; concurrent — messages from multiple nodes may arrive simultaneously.
- No direct response channel exists. To reply, initiate a `SendServiceOpsToNode()`.
- `messageType` / `payload` are service-defined opaque strings; format must be agreed upon between controller and service node implementations.

---

**`GetMetrics(networkID string) (map[string]string, error)`**

- Called after `Init()`; concurrent. Must return promptly.
- Return display-only metrics scoped to the given network. Values must be JSON-serializable.
- Maintain snapshots in background; do not compute on the call path.

---

**`Stop() error`**

- Idempotent. Must return promptly.
- Stop all background goroutines started in `Start()`.
- Return value is informational; the framework proceeds with shutdown regardless.

---

**`Finish()`**

- Called once, after `Stop()`, immediately before `.so` unload.
- Release all resources. Any goroutines still touching service-owned memory after `Finish()` returns cause undefined behavior.

---

### 5.2 ASNController (framework-provided)

All methods are goroutine-safe after `Init()` unless noted.

#### Resource Initialization

Must be called **in `Init()`**. All are one-shot per resource name; a second call returns an error.

| Method | Constraint | Returns |
|---|---|---|
| `InitLogger()` | Once total | `*log.Logger` (R/A/P/E loggers); see §12.4 |
| `InitDocDB(name)` | Once per `name` | Connected `DocDBHandler` |
| `InitTSDB(name)` | Once per `name` | Connected `TSDBHandler` |
| `InitLocker()` | Once total | Cluster-wide `Lock` |
| `GetIAM()` | Once total | `iam.Instance`; see §10 |
| `GetSubscription()` | Once total | `subscription.Instance`; see §11 |

---

#### Service Lifecycle Management

**`AddServiceToNode(nodeID string) error`**  
Loads the service `.so` onto the target node and triggers `Init()` on the node. Node must be online.

**`DeleteServiceFromNode(nodeID string) error`**  
Calls `Stop()` + `Finish()` on the node's service instance, then unloads the `.so`. Not a substitute for `StopService`.

**`StartService(scope ServiceScope, scopeList []string) error`**  
Triggers `Start(config)` on matched nodes.

**`StopService(scope ServiceScope, scopeList []string) error`**  
Triggers `Stop()` on matched nodes.

**`ResetService(scope ServiceScope, scopeList []string) error`**  
Triggers `Stop()` followed by `Start()` on matched nodes.

**`ServiceScope` / `scopeList` mapping:**

| Constant | Value | `scopeList` content |
|---|---|---|
| `ServiceScopeNetwork` | `1` | Network IDs |
| `ServiceScopeNetworkWithSubnetworks` | `2` | Network IDs (recursive) |
| `ServiceScopeNodeGroup` | `3` | Node Group IDs |
| `ServiceScopeNode` | `4` | Node IDs |

---

#### Ops Dispatch

**`SendServiceOps(scope ServiceScope, scopeList []string, opCmd, opParams string) (<-chan *OpsResponse, error)`**

- Re-entrant. Fan-out, asynchronous.
- If `error != nil`: invalid scope/scopeList; channel is nil.
- If `error == nil`: returns immediately. The framework dispatches to all matched nodes concurrently; responses stream into the channel, which is closed once all nodes have responded or timed out.

```go
resChan, paramErr := asnCtrl.SendServiceOps(scope, list, cmd, params)
if paramErr != nil { ... }
for res := range resChan {
    if res.FrameworkError != nil { ... }
    // res.ServiceResponse, res.ServiceError
}
```

**`SendServiceOpsToNode(nodeID, opCmd, opParams string) (*OpsResponse, error)`**

- Re-entrant. Point-to-point, synchronous. Blocks until the node responds or the framework timeout fires.
- If `error != nil`: invalid `nodeID`.

---

#### Config Ops Dispatch

These methods persist and fan out config ops to service nodes. They are the framework-provided counterpart to the `ASNServiceController.AddConfigOps` / `UpdateConfigOp` / `DeleteConfigOps` callbacks (§5.1). Config ops scope is limited to `ServiceScopeNodeGroup` (3) or `ServiceScopeNode` (4) only.

**`AddConfigOps(serviceScope ServiceScope, scopeID string, configParams []string) (<-chan *OpsResponse, error)`**

- Persists the new config ops for the given scope, then fans out to all affected service nodes concurrently. Return semantics are identical to `SendServiceOps`: if `error != nil`, the scope/scopeID is invalid and the channel is nil; otherwise responses stream into the channel, which closes after all nodes have responded or timed out.
- Each `OpsResponse` reflects the result of calling `ASNService.AddConfigOps` on the node.

**`UpdateConfigOp(serviceScope ServiceScope, scopeID, configOpID, configParam string) (<-chan *OpsResponse, error)`**

- Updates a single existing config op (identified by `configOpID`) for the given scope, persists the change, and fans out to affected nodes.

**`DeleteConfigOps(serviceScope ServiceScope, scopeID string, configOpIDs []string) (<-chan *OpsResponse, error)`**

- Deletes config ops by ID for the given scope, persists, and fans out to affected nodes.

**`ListConfigOps(serviceScope ServiceScope, scopeID string) ([]ConfigOp, error)`**

- Returns config ops **directly attached** to the given scope. Does not traverse the group-to-node inheritance hierarchy. Synchronous; does not fan out to nodes.

---

#### Node Topology

**`GetNetworks() ([]*Network, error)`**  
Returns the full network tree. Each `Network` embeds nested `Networks` (subnetworks).

**`GetNodeByID(nodeID string) (*Node, error)`**  
Returns full node details: hardware info, service-defined `Metadata`, and `ServiceInfo` (state, config source, active config ops).

**`UpdateNodeMetadata(nodeID, metadata string) error`**  
Persists an opaque service-defined string on the node. Retrievable via `GetNodeByID().Metadata`.

**`SetConfigOfNode(nodeID, config string) error`**  
Persists the service config (YAML/UTF-8) for the node, used on the next `StartService` call.

**`GetNodesOfNetwork(networkID string, withService bool) ([]*Node, []*Link, error)`**
- `withService = true` filters to nodes that have successfully loaded and initialized this service (state ≥ `ServiceStateInitialized`).
- Internal links: both endpoints within the network; the `To` node is present in the returned `nodes` slice.
- External links: `To` endpoint is outside the network; it is not included in `nodes`.

**`SubscribeNodeStateChanges() (<-chan *NodeStateChange, error)`**

- **One-shot.** A second call returns an error.
- On subscription: delivers a `NodeStateChange` for every node's current state (initial snapshot), then delivers incremental changes.
- The channel is never closed during normal framework operation.

---

#### Node Group Management

All methods are re-entrant.

| Method | Effect |
|---|---|
| `CreateNodeGroup(networkID, name, desc, metadata)` | Creates a group scoped to this service within the network |
| `ListNodeGroups(networkID)` | Lists all groups for this service in the network |
| `GetNodeGroupByID(nodeGroupID)` | Returns group details including `Metadata` and active `ConfigOps` |
| `UpdateNodeGroupMetadata(nodeGroupID, metadata)` | Persists service-defined metadata on the group |
| `DeleteNodeGroup(nodeGroupID)` | Deletes the group; member nodes are not affected |
| `SetConfigOfNodeGroup(nodeGroupID, config)` | Persists service config for the group; inherited by member nodes unless overridden |
| `AddNodesToNodeGroup(nodeGroupID, nodeIDs)` | Adds nodes to the group |
| `RemoveNodesFromNodeGroup(nodeGroupID, nodeIDs)` | Removes nodes from the group |

---

## 6. Service Node Side — `snapi`

### 6.1 ASNService (service-implemented)

#### Lifecycle Sequence

```
StaticResource()              any time, including pre-Init; re-entrant
  ├─ Version()
  └─ SharedData()

Init(asnServiceNode)          once; not re-entrant
Start(config)                 after Init; sequential; may be called multiple times
                              → returns runtimeErrChan

  ├─ ApplyServiceOps()        after Start; explicitly concurrent
  ├─ AddConfigOps()           after Start; concurrent
  ├─ UpdateConfigOp()         after Start; concurrent
  ├─ DeleteConfigOps()        after Start; concurrent
  ├─ OnQuerySharedData()      after Init; concurrent
  └─ OnSubscribeSharedData()  after Init; concurrent

Stop()                        idempotent
Finish()                      once; after Stop
```

#### Method Contracts

---

**`StaticResource() StaticResource`**

- Pre-`Init()` safe; re-entrant. Returns a stable object.

---

**`StaticResource.Version() commonapi.Version`**

- Returns a hardcoded version constant.

---

**`StaticResource.SharedData() (aggregated, subscribable []string)`**

- Declares the shared data keys this service provides to other services on the same node.
- `aggregated`: keys available via `QueryServiceSharedData` (pull model).
- `subscribable`: keys available via `SubscribeServiceSharedData` (push/stream model).
- Return `(nil, nil)` if this service does not provide shared data.
- The framework uses these declarations to validate inbound requests from consuming services.

---

**`Init(asnServiceNode ASNServiceNode) error`**

- Called once after `.so` load. Not re-entrant. Same contract as controller-side `Init()`.
- All `Init*` calls on `asnServiceNode` must occur here.

---

**`Start(config string) (runtimeErrChan <-chan error, err error)`**

- Called after `Init()` and on every subsequent config change. Sequential; may be called multiple times.
- **Must be idempotent.** Each invocation must fully supersede the previous config.
- If the service cannot hot-reload, return `ErrRestartNeeded`. The framework will execute `Stop()` → `Start()` with the new config (no re-initialization).
- `runtimeErrChan`: used to report fatal runtime errors discovered **after** `Start()` returns. Only unrecoverable conditions should be sent; the framework transitions the service to `Malfunctioning` on receipt.
- Synchronous error (non-`ErrRestartNeeded`) → `Malfunctioning`.

---

**`ApplyServiceOps(opCmd, opParams string) (resp string, err error)`**

- Called after `Start()`. **Explicitly concurrent** — the framework may invoke this simultaneously from multiple goroutines. All shared state access must be synchronized.
- `opCmd` / `opParams` are service-defined; both controller and service node must agree on per-command schemas.
- `resp` → `OpsResponse.ServiceResponse`; `err` → `OpsResponse.ServiceError`. Neither represents a fatal error; use `runtimeErrChan` for unrecoverable failures.
- The framework enforces a per-call timeout. Must return promptly.

---

**`AddConfigOps(configParams []string) (resp string, err error)`**  
**`UpdateConfigOp(oldConfigParam, newConfigParam string) (resp string, err error)`**  
**`DeleteConfigOps(configParams []string) (resp string, err error)`**

- Called after `Start()`, potentially concurrent. Guard shared state.
- Returning an error from any of these methods sets the service state to `Malfunctioning`. Only return errors for genuinely unrecoverable failures.

---

**`OnQuerySharedData(keys []string) (map[string]any, error)`**

- Called when another service on the same node calls `QueryServiceSharedData()` targeting this service.
- Called after `Init()`; concurrent.
- Implement only if this service is a data provider (i.e., `aggregated` keys declared in `SharedData()`). Return `ErrKeyNotFound` for unknown keys.

---

**`OnSubscribeSharedData(keys []string) (map[string]<-chan any, error)`**

- Called when another service calls `SubscribeServiceSharedData()` targeting this service.
- Called after `Init()`; concurrent.
- Return a dedicated channel per key. **Each channel must be closed** after all values are sent to signal end-of-stream. Return `ErrKeyNotFound` for unknown keys.

---

**`Stop() error`**

- Idempotent. Must return promptly. Any returned error may influence service state.

---

**`Finish()`**

- Called once after `Stop()`, before `.so` unload. Same contract as controller-side `Finish()`.

---

### 6.2 ASNServiceNode (framework-provided)

All methods are goroutine-safe after `Init()`.

#### Resource Initialization

One-shot per resource name; called in `Init()`.

| Method | Constraint | Returns |
|---|---|---|
| `InitLogger()` | Once total | `*log.Logger` |
| `InitDocDB(name)` | Once per `name` | Connected `DocDBHandler` |
| `InitTSDB(name)` | Once per `name` | Connected `TSDBHandler` |

---

#### Node Information

**`GetNodeType() commonapi.NodeType`**  
Returns the hardware/role type of this node. See §13 for the full enum.

**`GetNodeInfo() *snapi.NodeInfo`**  
Returns `snapi.NodeInfo`, which embeds `commonapi.NodeInfo` (interfaces, IPMI, management, hardware params) plus node `ID` and the active `ConfigOps` string list.

---

#### Communication

**`SendMessageToController(messageType, payload string) error`**

- Fire-and-forget upcall to the controller, handled by `ASNServiceController.HandleMessageFromNode()`.
- No response is returned. To receive a reply, the controller must initiate a separate `SendServiceOpsToNode()`.

---

#### Cross-Service Data Access

**`GetSharedData(serviceName string) (aggregated, subscribable []string, error)`**  
Queries the key registry of another service on this node. Returns `ErrServiceNotFound` if the service is not loaded or is not in `Running` state.

**`QueryServiceSharedData(serviceName string, keys []string) (map[string]any, error)`**  
Pull-model fetch. Returns `ErrServiceNotFound` or `ErrKeyNotFound` on failure.

**`SubscribeServiceSharedData(serviceName string, keys []string) (map[string]<-chan any, error)`**  
Push-model subscription. Returns one channel per key, closed by the provider when the stream ends.  
**Constraint:** only one active subscription per key is permitted. Opening a second subscription before the first channel closes is undefined behavior.

---

## 7. Config Ops

Config Ops are persistent, incremental configuration directives stored per node or node group, complementing the monolithic YAML config passed to `Start()`.

### Data Flow

```
Controller                    Framework                 Service Node
──────────                    ─────────                 ────────────
ASNController.AddConfigOps()
  ├─ persist ops ────────────────────────────────────►
  │                                                    ASNService.AddConfigOps()
  └─ OpsResponse chan ◄────────────────────────────────
ASNServiceController.AddConfigOps()  ← controller-side hook
```

The framework persists and dispatches to nodes **before** invoking `ASNServiceController.AddConfigOps()`.

### Scoping

- Config Ops are scoped to `ServiceScopeNodeGroup` (3) or `ServiceScopeNode` (4) only. Network-level scope is not supported.
- Group-level ops apply to all member nodes unless the node carries its own direct ops.
- `ListConfigOps(scope, scopeID)` returns ops **directly attached** to the given scope only — it does not traverse the group-to-node hierarchy.
- `ConfigOp.ID` is framework-assigned. Use it to identify ops in update and delete calls.
- `ConfigOp.ConfigParams` is an opaque service-defined string.

### Error Semantics

A non-nil error from any config op method on `ASNService` immediately transitions the node's service state to `Malfunctioning`. Return errors only for unrecoverable conditions; internal validation failures should be handled within the service and reported via metrics or logs.

---

## 8. Ops Commands

Ops commands are ephemeral, on-demand directives from the controller to service nodes. They are not persisted.

### Calling Modes

| Method | Dispatch | Blocking | Scope |
|---|---|---|---|
| `SendServiceOps()` | Fan-out | No (channel) | Multiple nodes via `ServiceScope` |
| `SendServiceOpsToNode()` | Point-to-point | Yes | Single node by ID |

### Message Protocol

`opCmd` is a service-defined command identifier. `opParams` is a service-defined payload (typically JSON). Controller and service node implementations must share the schema for each `opCmd`.

```go
// Controller: dispatch
params, _ := json.Marshal(MyOpParams{...})
resChan, paramErr := asnCtrl.SendServiceOps(scope, list, "cmd-name", string(params))

// Service node: handle in ApplyServiceOps
var p MyOpParams
json.Unmarshal([]byte(opParams), &p)
```

### Response Structure

```
Controller              Framework              Node(s)
──────────              ─────────              ──────
SendServiceOps() ──────► (fan-out) ───────────► ApplyServiceOps()
     │                                               │ resp, err
     ◄── resChan: *OpsResponse per node ◄───────────
         (closed after all nodes respond or timeout)
```

`OpsResponse` fields:

| Field | Content |
|---|---|
| `NodeID` | Responding node |
| `Timestamp` | Response time |
| `FrameworkError` | Non-nil if the framework could not reach or invoke the service |
| `ServiceResponse` | `resp` from `ApplyServiceOps()`; valid only when `FrameworkError == nil` |
| `ServiceError` | `err` from `ApplyServiceOps()`; valid only when `FrameworkError == nil` |

---

## 9. Cross-Service Data Sharing

Services co-located on the same service node can exchange data through the framework's local registry. The framework routes calls without network overhead.

### Models

| Model | Consumer API | Provider callback | Use case |
|---|---|---|---|
| Query (pull) | `QueryServiceSharedData` | `OnQuerySharedData` | Point-in-time snapshot fetch |
| Subscribe (push) | `SubscribeServiceSharedData` | `OnSubscribeSharedData` | Continuous / time-series stream |

### Provider Contract

- Declare all provided keys in `StaticResource.SharedData()` before `Init()`.
- Implement `OnQuerySharedData` for `aggregated` keys; implement `OnSubscribeSharedData` for `subscribable` keys.
- For subscriptions, close each returned channel after all values are emitted.
- Return `ErrKeyNotFound` for keys not in your declared set.

### Consumer Contract

- Call `GetSharedData(serviceName)` to discover available keys before consuming.
- For subscriptions: only one active subscription per key is permitted at any time. Re-subscribe only after the previous channel is closed.

### Key Negotiation

Key names are service-defined and opaque to the framework. Producing and consuming services must agree on names and value types out-of-band, documented in their respective service API contracts.

---

## 10. IAM

Obtained via `ASNController.GetIAM()`. `iam.Instance` is goroutine-safe. All operations are scoped to the calling service's IAM namespace; accounts, groups, and accesses created by one service are not visible to another.

### Service Namespace and Admin Group

Each service operates in its own isolated IAM namespace. In addition, the framework automatically provisions one **admin group** per service (named `{serviceName}-admin`) in the ASN management namespace. This group cannot be managed via the service-facing `Group*` API — it is maintained exclusively by the framework. When account queries return an `Account`, the `ServiceAdmin` field reflects whether the account is a current member of this admin group.

### Account

Every account returned by IAM calls is represented as an `Account` struct:

| Field | Type | Description |
|---|---|---|
| `ID` | `string` | UUID |
| `Username` | `string` | Login username |
| `UsernameModified` | `bool` | Whether the username has been changed from its initial value |
| `Metadata` | `string` | Service-defined opaque string |
| `DeviceLimit` | `int` | Maximum number of concurrent logged-in devices |
| `Password` | `bool` | Whether a password credential is set |
| `Phone` | `Phone{CountryCode, Number}` | Bound phone number |
| `Email` | `string` | Bound email address |
| `Totp` | `bool` | Whether a TOTP authenticator is enrolled |
| `MfaEnabled` | `bool` | Whether MFA is enabled for this account |
| `WeChat` | `*AccountWeChatInfo` | WeChat binding info (Bound, Nickname, HeadImgURL) |
| `Apple` | `*AccountAppleInfo` | Apple binding info (Bound, Email, EmailVerified, IsPrivateEmail) |
| `Google` | `*AccountGoogleInfo` | Google binding info (Bound, Email, EmailVerified, Name, Picture) |
| `Passkeys` | `[]AccountPasskey` | Enrolled passkeys (ID, DeviceID, CredentialID) |
| `ServiceAdmin` | `bool` | Whether this account is in the service's admin group (managed by the framework) |
| `Groups` | `[]string` | Group names (within this service's namespace) the account belongs to |
| `Devices` | `map[string]*Device` | Active devices keyed by device ID |
| `TimeInfo` | `TimeInfo{CreatedAt, UpdatedAt}` | Timestamps |

### Account Management

| Category | Methods |
|---|---|
| CRUD | `AccountCreate`, `AccountDelete`, `AccountGet`, `AccountList`, `AccountListByIDs`, `AccountExists` |
| Profile update | `AccountRename`, `AccountMetadataUpdate`, `AccountPhoneUpdate`, `AccountEmailUpdate` |
| Third-party binding | `AccountWeChatUpdate`, `AccountAppleUpdate`, `AccountGoogleUpdate` |
| Password | `AccountPasswordUpdate` (user-initiated, requires old password), `AccountPasswordReset` (admin-initiated, no old password required) |

**`AccountCreate`** signature:

```go
AccountCreate(
    username, metadata string,
    password string,
    email, emailCode string, skipEmailValidation bool,
    phone *Phone, phoneCode string, skipPhoneValidation bool,
) (accountID string, err error)
```

Fields that are empty or nil are optional. `skipEmailValidation` / `skipPhoneValidation` bypass OTP pre-verification for admin-initiated creation.

**`AccountGet(accountID, username, countryCode, number, email string) (*Account, error)`** — any single non-empty field is used as the lookup key; the rest are ignored.

**`AccountList(username, countryCode, number, email string) ([]*Account, error)`** — filter parameters; empty strings match all.

### OTP Code Sending

Before OTP-based login or MFA verification, send a code to the user:

```go
AccountPhoneSend(countryCode, number string, sendWhenExistOnly bool) (sessionToken string, err error)
AccountEmailSend(email string, sendWhenExistOnly bool) (sessionToken string, err error)
```

`sendWhenExistOnly = true` silently succeeds (no code sent) if no account with the given contact exists, preventing enumeration.

### Authentication

Use `LoginMethods()` to query which methods are enabled at runtime:

```go
LoginMethods() (
    usernameAndPassword, emailAndPassword, phoneAndPassword,
    emailCode, phoneCode,
    weChat, apple, google, passkey bool,
    err error,
)
```

| Method | Entry point |
|---|---|
| Username + password | `LoginOrCreateWithPassword` |
| Email + password | `LoginOrCreateWithPassword` (pass email, leave username empty) |
| Phone + password | `LoginOrCreateWithPassword` (pass countryCode+number, leave username empty) |
| Phone OTP | `LoginOrCreateWithPhone` |
| Email OTP | `LoginOrCreateWithEmail` |
| WeChat OAuth | `LoginOrCreateWithWeChat` |
| Apple Sign-In | `LoginOrCreateWithApple` |
| Google Sign-In | `LoginOrCreateWithGoogle` |
| Passkey (WebAuthn) | `AccountPasskeyLoginChallengeGet` → `AccountPasskeyAuth` |

All `LoginOrCreate*` and `AccountPasskeyAuth` share this signature pattern:

```go
(
    device *DeviceInfo,
    userClaims string,
    durationAccess, durationRefresh time.Duration,
    ... method-specific credentials ...,
    createIfNotExist bool,
) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
```

- `device`: identifies the calling device; required. See §Device Management.
- `userClaims`: opaque service-defined string embedded in the issued token; retrievable via `TokenVerify`.
- `durationAccess` / `durationRefresh`: token expiry durations.
- `createIfNotExist`: if `true`, an account is created on first successful credential validation.
- When `needMfa == true`, `tokenSet` is a pre-MFA token. The session is not fully authorized until `MFALoginVerify` succeeds.

**`PasswordVerify(username, countryCode, number, email, password string) error`** — validates credentials without issuing tokens. Useful for re-authentication flows.

### Session Management

**`Logout(accountID, deviceID string) error`** — invalidates the session for a specific device and revokes its tokens.

**`AppleRedirect(w http.ResponseWriter, r *http.Request)`** — handles the Apple OAuth redirect callback. Mount this on the service's web router at the Apple-configured redirect URI.

### Token Lifecycle

| Method | Behavior |
|---|---|
| `TokenVerify(accessToken string)` | Returns `(mfaNeeded, accountID, username, deviceID, userClaims, err)` |
| `TokenRefresh(userClaims, tokenSet, durationAccess)` | Issues a new access token from a valid refresh token |
| `TokenRevoke(accessToken string)` | Immediately invalidates the access token |
| `JWKSGet() (string, error)` | Returns the JSON Web Key Set (JWKS) for external token verification |

When `mfaNeeded == true` from `TokenVerify`, the token is a pre-MFA token; only MFA endpoints should accept it until `MFALoginVerify` upgrades it to a fully-authorized session token.

### MFA

**Per-account MFA control:**

```go
AccountEnableMFA(accountID string) error   // enables MFA requirement for the account
AccountDisableMFA(accountID string) error  // disables MFA requirement
```

**Enrollment and verification:**

| Type | Constant | Enrollment / verification flow |
|---|---|---|
| TOTP | `MfaTypeTotp` | `TotpBind(accountID)` → share QR/secret → `TotpBindConfirm(accountID, code)`; unbind: `TotpUnbind`; verify: `MFALoginVerify` |
| Email OTP | `MfaTypeEmail` | `AccountEmailSend(email, false)` (send code); verify: `MFALoginVerify` |
| Phone OTP | `MfaTypeSms` | `AccountPhoneSend(countryCode, number, false)` (send code); verify: `MFALoginVerify` |
| Passkey | `MfaTypePasskey` | `AccountPasskeyLoginChallengeGet(domain)` → `MFALoginVerify` |

**`TotpBind(accountID string) (img, issuer, secret string, err error)`** — returns a QR code image (data URI), issuer name, and raw TOTP secret for display to the user.

**`MFALoginVerify(accessToken string, method MfaType, code, domain, sessionID, data string) (*TokenSet, error)`** — upgrades a pre-MFA access token to a fully-authorized token. The `code`, `domain`, `sessionID`, `data` parameters are method-specific; unused fields for a given method should be empty strings.

**Service-wide MFA enforcement:**

```go
ServiceMfaSet(mfaRequired bool) error  // enforce or relax MFA for all accounts in this service
MfaEnforced() bool                     // reflects the current enforcement state
```

### Passkey (WebAuthn)

```
Bind:   AccountPasskeyBindChallengeGet(domain, accountID) → sessionID, data
        AccountPasskeyBind(domain, accountID, sessionID, deviceID, data)

Login:  AccountPasskeyLoginChallengeGet(domain) → sessionID, data
        AccountPasskeyAuth(device, userClaims, durationAccess, durationRefresh, domain, sessionID, data)

Unbind: AccountPasskeyUnbind(accountID, passkeyID)   // remove a single passkey
        AccountPasskeyUnbindAll(accountID)            // remove all passkeys for the account
```

### Device Management

Every login call requires a `*DeviceInfo`. The framework tracks one `Device` record per (account, device) pair.

**`DeviceInfo` fields:**

| Field | Type | Values |
|---|---|---|
| `Category` | `DeviceCategory` | `Phone`, `Tablet`, `Wearable`, `Browser`, `MiniProgram`, `PC` |
| `Type` | `DeviceType` | Per-category sub-type (e.g., `PhoneIPhone`, `PhoneAndroid`, `TabletIPad`, `PCMacOS`) |
| `OS` | `DeviceOs` | `ios`, `android`, `watchos`, `android_wear`, `wechat`, `windows`, `macos`, `linux` |
| `Language` | `DeviceLanguage` | `EN`, `ZH` |
| `Name` | `string` | User-visible device name |
| `Model` | `string` | Hardware model string |
| `SerialNumber` | `string` | Device serial number |
| `PushToken` | `string` | Notification push token |
| `Metadata` | `string` | Service-defined opaque string |

A `Device` embeds `DeviceInfo` and adds an `ID string` (UUID assigned by the framework on first login).

| Method | Effect |
|---|---|
| `DeviceLimitUpdate(accountID, limit int)` | Sets the maximum number of concurrent active devices for the account |
| `DeviceInfoUpdate(accountID string, device *Device)` | Updates the stored device record (push token, model, metadata, etc.) |
| `DeviceDelete(accountID, deviceID string)` | Removes the device and revokes all its associated tokens |

### Groups

Groups are service-scoped collections of accounts. They are used to assign accesses in bulk.

```
Group ──has members──► Account(s)
  └──has accesses──► Access(name, scope, operation, TimeControl)
```

| Method | Effect |
|---|---|
| `GroupCreate(groupName, metadata string)` | Creates a group within this service's namespace |
| `GroupDelete(groupName string)` | Deletes the group |
| `GroupExists(groupName string)` | Checks existence |
| `GroupRename(oldName, newName string)` | Renames the group |
| `GroupMetadataUpdate(groupName, metadata string)` | Updates group metadata |
| `GroupGet(groupName string)` | Returns `*Group{GroupName, GroupMembers, Metadata, TimeInfo}` |
| `GroupList()` | Lists all groups in this service's namespace |
| `GroupMemberList(groupName string)` | Returns all `*Account` members of the group, with `ServiceAdmin` populated |
| `AccountJoinGroup(groupName string, accountIDs []string)` | Adds accounts to the group |
| `AccountLeaveGroup(groupName string, accountIDs []string)` | Removes accounts from the group |
| `AccountGroupList(accountID string)` | Lists all groups (within this service's namespace) the account belongs to |

> The framework-managed admin group (`{serviceName}-admin`) is **not** included in `GroupList()` results and cannot be manipulated via the `Group*` API.

### Accesses and Access Control

An **Access** defines a named, time-controlled permission rule:

| Field | Description |
|---|---|
| `AccessName` | Unique name within the service namespace |
| `Scope` | The resource being protected (e.g., the service name, or a sub-resource) |
| `Operation` | Permitted actions: `"view"`, `"manage"`, or `"view;manage"` |
| `Time` | `*TimeControl`; `nil` means always valid |

| Method | Effect |
|---|---|
| `AccessCreate(name, scope, operation string, time *TimeControl)` | Defines a new access rule |
| `AccessUpdate(name, scope, operation string, time *TimeControl)` | Updates an existing access rule |
| `AccessDelete(name string)` | Removes the access rule |
| `AccessExists(name string)` | Checks existence |
| `AccessList()` | Lists all access rules in this service's namespace |
| `AccessGrantToGroup(groupName string, accesses []string)` | Grants the named access rules to a group |
| `AccessRevokeFromGroup(groupName string, accesses []string)` | Revokes named access rules from a group |
| `GroupAccessList(groupName string)` | Lists access rules currently granted to a group |
| `AccountAccessList(accountID string)` | Returns all effective access rules for the account, as `map[serviceName][]*Access` where the map key is the service name whose namespace the access belongs to (e.g., this service's name, or `"asn"` for framework-level accesses such as the admin group access) |

### TimeControl

`TimeControl` defines when an access rule is valid. `nil` means always valid.

| Field | Type | Description |
|---|---|---|
| `TimeRanges` | `[]TimeRange{Start, End}` | One or more clock-time windows within each repeat period |
| `RepeatFrequency` | `RepeatFrequency` | `OnlyOnce` (0), `Daily` (1), `Weekly` (2), `Monthly` (3) |
| `RepeatEndTime` | `time.Time` | When repetition stops; zero means no end |
| `RepeatInterval` | `int` | Every N periods (e.g., 2 = every other week) |
| `RepeatIndexes` | `[]int` | For `Weekly`: day-of-week (1–7); for `Monthly`: day-of-month (1–31) |
| `IgnoreLoc` | `bool` | If `true`, time comparisons ignore timezone |

Examples:
- Every weekday: `RepeatFrequency=Weekly, RepeatInterval=1, RepeatIndexes=[1,2,3,4,5]`
- Every 2 months on the 1st and 15th: `RepeatFrequency=Monthly, RepeatInterval=2, RepeatIndexes=[1,15]`

### System Configuration

| Method | Effect |
|---|---|
| `RenameEnabled() bool` | Returns whether the IAM deployment allows username changes |
| `ServiceMfaSet(mfaRequired bool) error` | Enforces or relaxes service-wide MFA |
| `MfaEnforced() bool` | Current service-wide MFA enforcement state |
| `JWKSGet() (string, error)` | Returns the JWKS document for external token verification |

### Phone Country Code Filtering

`SupportedCountryCodesGet() (PhoneCountryCodeMode, []string)` returns:
- Mode `"all"` — no restriction; the list is ignored
- Mode `"include"` — only the listed country codes are accepted
- Mode `"exclude"` — the listed country codes are rejected

Use this to validate phone numbers before calling login or OTP-send methods.

---

## 11. Subscription

Obtained via `ASNController.GetSubscription()`. Supports Apple App Store, Google Play, and Stripe.

### Platform Registration

Each platform is registered independently during `Start()`. Each returns an HTTP handler function to be mounted on the service's web router and an error channel for backend failures.

```go
appleWebhookFn, errChan, err := sub.AddApple(envConfig, apiConfig)
googleWebhookFn, errChan, err := sub.AddGoogle(envConfig, replayConfig)
stripeWebhookFn, errChan, err := sub.AddStripe(config)
```

Monitor each `errChan` in a background goroutine. Mount the returned handlers via `WebHandler()`.

### Notifications

`GetNotificationChannel()` returns a unified channel that emits a string notification on any subscription lifecycle event (new, renew, cancel, etc.) from any registered platform.

### Subscription Queries

| Method | Behavior |
|---|---|
| `GetUserSubscription(accountID)` | Returns active subscription; `(nil, false, nil)` if none |
| `ListUserSubscriptions()` | Returns all active subscriptions |
| `DeleteUserSubscription(accountID)` | Removes the local record; does not cancel on the platform |
| `RestoreApplePurchaseToken(accountID, token)` | Re-links an existing App Store purchase to an account |
| `RestoreGooglePurchaseToken(accountID, token)` | Re-links an existing Play Store purchase to an account |

### Stripe-Specific

| Method | Returns |
|---|---|
| `GetStripeProductInfo(priceID)` | `*Product` with name, description, and multi-currency pricing |
| `GetStripePaymentLink(accountID, priceID, quantity, redirectURL)` | Stripe Checkout URL |
| `GetStripeBillingPortalUrl(accountID, returnURL)` | Stripe Customer Portal URL |

---

## 12. Common Abstractions

### 12.1 Document Database (DocDB)

`DocDBHandler` (returned by `InitDocDB`) is compatible with MongoDB and file-based backends.

```
DocDBHandler
  ├── FindOrCreateCollection(name, options) → DocCollection
  ├── DeleteCollection(name)
  ├── ListCollections()
  ├── CreateFile(fileName, metadata, file)
  └── FindFiles(queryString)
```

Index creation: pass `DocCollOptions.Indexes` to `FindOrCreateCollection`. Key format for compound indexes: `"col1;col2;..."`.

`DocCollection` operations:

| Method | Cardinality contract |
|---|---|
| `AddRecord(record)` / `AddRecords(records)` | — |
| `FindRecord(queryJson, fieldFilter)` | Errors if 0 or >1 match |
| `FindRecords(queryJson, page, num, sorting, fieldFilter)` | Any cardinality |
| `FindRecordWithRegex` / `FindRecordsWithRegex` | Same as above, regex query |
| `UpdateRecord(queryJson, newRecord)` | Errors if 0 or >1 match |
| `UpdateRecords(queryJson, newRecord)` | Any cardinality |
| `DeleteRecord(queryJson)` | Errors if 0 or >1 match |
| `DeleteRecords(queryJson)` | Any cardinality |
| `ArrayAppend(queryJson, newRecord, ignoreDuplicate)` | Appends to an array field |
| `ArrayDelete(queryJson, newRecord)` | Removes from an array field |
| `ListAllRecord()` | Returns all documents |

**Pagination semantics** (`page`, `num` parameters):

| `page` | `num` | Behavior |
|---|---|---|
| `<= -1` | `<= 0` | No limit, no pagination |
| `<= -1` | `> 0` | Limit only |
| `>= 0` | `> 0` | Paginated; `page` is 0-indexed |
| `>= 0` | `<= 0` | Error |

**Type conversion helpers** (`commonapi`):

| Function | Direction |
|---|---|
| `MapToStruct[T](map)` / `MapsToStructs[T]([]map)` | `map[string]any` → typed struct (via JSON) |
| `StructToMap[T](struct)` / `StructsToMaps[T]([]struct)` | typed struct → `map[string]any` (via JSON) |

---

### 12.2 Time-Series Database (TSDB)

`TSDBHandler` (returned by `InitTSDB`) is compatible with InfluxDB v1/v2 and MongoDB time-series collections.

```
TSDBHandler
  ├── FindOrCreateCollection(name, options) → TSCollection
  ├── DeleteCollection(name) / ListCollections()
  ├── AddRetentionPolicies(rps) / DeleteRetentionPolicies(rps)
  └── NewPoint() / AddDataPoint(pt, useTimeout) / AddDataPoints(pts, useTimeout)
```

`TSCollection` operations:

| Method | Notes |
|---|---|
| `NewPoint()` / `NewQuery()` | Builder pattern |
| `AddDataPoint(pt)` / `AddDataPoints(pts)` | Write |
| `FindDataPoints(query)` → `TSQueryResult` | Typed query builder |
| `FindDataPointsBySyntax(syntax)` | Raw database syntax |
| `FindDataPointsWithRegex(query)` | Regex-filtered read |
| `FindPivotDataPoints(query)` → `TSPivotQueryResult` | Row-pivoted result (tag groups as columns) |
| `CountDataPoints(query)` | Count without data retrieval |

`TSQueryResult`: `Results map[string]TSMatrix` where `TSMatrix` maps tag-group strings to `TSSamples` (ordered `[]TSSamplePoint{Timestamp, Value}`).

---

### 12.3 Distributed Lock

Returned by `ASNController.InitLocker()`. Available on the **controller side only** (no locker on service nodes).

```go
type Lock interface {
Lock(key, identifier string) error
Unlock(key, identifier string) error
}
```

`key` is the resource name. `identifier` is the lock-holder identity (e.g., operation ID or node ID); it prevents unintended unlocks by a different holder.

---

### 12.4 Logger

Returned by `InitLogger()` on both controller and service-node sides.

| Field | Alias | Purpose |
|---|---|---|
| `Logger.R` | `rlog` / `rtlog` | Runtime — general operational events |
| `Logger.A` | `alog` / `apilog` | API access — gRPC, REST, CLI calls |
| `Logger.P` | `plog` / `perflog` | Performance — latency, throughput |
| `Logger.E` | `elog` / `entitylog` | Entity audit — create / read / update / delete |

Each implements `ASNLogger`: `Trace`, `Debug`, `Info`, `Warn`, `Error`, `Fatal`, `Panic` (both plain and `f`-format variants).

---

### 12.5 Version

Format: `vMAJOR.MINOR[.BUILD[-SUFFIX]]`

```go
v, err := commonapi.InitVersion("v26.5.4")
v.ToString()       // → "v26.5.4"
v.Compare(other)   // VersionCompareGreater(1) | VersionCompareLess(2) | VersionCompareEqual(3)
```

---

## 13. Network and Node Topology

### Network Hierarchy

`GetNetworks()` returns top-level networks. Each `Network` embeds `Networks []Network` (recursive subnetworks), linked via `ParentID`.

`Network.Tiers` is a subset of the 12-level location hierarchy:

```
world → country → state → city → district → campus →
building → floor → room → row → rack → unit
```

### Node Types (`commonapi.NodeType`)

| Constant | Value |
|---|---|
| `NodeTypeRouter` | `"router"` |
| `NodeTypeSwitch` | `"switch"` |
| `NodeTypeAppliance` | `"appliance"` |
| `NodeTypeFirewall` | `"firewall"` |
| `NodeTypeLoadBalancer` | `"lb"` |
| `NodeTypeAccessPoint` | `"ap"` |
| `NodeTypeDevice` | `"device"` |
| `NodeTypeServer` | `"server"` |

### Node Hardware Info (`commonapi.NodeInfo`)

| Field | Type | Content |
|---|---|---|
| `Interfaces` | `map[string]*Interface` | Per-interface IP and role tags |
| `Ipmi` | `*Ipmi` | OOB management credentials |
| `Management` | `*Management` | Hostname and management IP |
| `DeviceInfo` | `*DeviceInfo` | Serial number, vendor, model |
| `DeviceParam` | `*DeviceParam` | CPU cores, memory (bytes), disk (bytes) |

`Interface.Tags` (`NetIfType`): `data` · `control` · `management` · `inbound` · `outbound`

### Links (`capi.Link`)

```go
type Link struct {
ID          string    // UUID
Description string
Bandwidth   int64     // bps; symmetric
From, To    *LinkNode // {NodeID, Interface}
}
```

Internal links: both endpoints within the queried network; `To` node is included in the `GetNodesOfNetwork` result.  
External links: `To` endpoint is outside the network; not included in the result.

---

## 14. Implementation Checklist

### ASNServiceController

- [ ] `StaticResource()` and all sub-methods are pre-`Init()` safe and re-entrant
- [ ] `CLICommands()` is purely declarative — no side effects, no goroutines
- [ ] `WebHandler()` defers all resource initialization to `Init()`
- [ ] `Init()` calls each `Init*` / `Get*` method exactly once and stores the `ASNController` handle
- [ ] `Start()` returns promptly; all background work runs in goroutines managed by `Start()`
- [ ] `Start()` fully supersedes prior config on each invocation
- [ ] `HandleMessageFromNode()` synchronizes access to shared state (concurrent)
- [ ] Config op callbacks (`AddConfigOps`, `UpdateConfigOp`, `DeleteConfigOps`) synchronize shared state
- [ ] `GetMetrics()` returns promptly from pre-computed snapshots
- [ ] `Stop()` is idempotent and returns promptly; all goroutines from `Start()` are stopped
- [ ] `Finish()` releases all resources; no goroutines remain after return

### ASNService

- [ ] `StaticResource.SharedData()` declares all provided keys accurately, or returns `(nil, nil)`
- [ ] `Init()` calls each `Init*` method exactly once and stores the `ASNServiceNode` handle
- [ ] `Start()` is idempotent; returns `ErrRestartNeeded` if hot-reload is unsupported
- [ ] `runtimeErrChan` is used only for unrecoverable post-start failures
- [ ] `ApplyServiceOps()` uses explicit synchronization on all shared state (explicitly concurrent)
- [ ] Config op callbacks return errors only for unrecoverable conditions
- [ ] `OnQuerySharedData()` returns `ErrKeyNotFound` for undeclared keys
- [ ] `OnSubscribeSharedData()` closes every returned channel after all values are sent
- [ ] `Stop()` is idempotent and returns promptly
- [ ] `Finish()` releases all resources; no goroutines remain after return

### Cross-Cutting

- [ ] `SubscribeNodeStateChanges()` is called at most once
- [ ] `runtimeErrChan` carries only fatal, unrecoverable errors — not service-internal errors
- [ ] `opCmd` / `opParams` schemas are formally defined and shared between controller and service node
- [ ] Shared data key names and value types are documented and agreed upon between service teams
- [ ] Config op payload formats are versioned for compatibility across rolling upgrades
