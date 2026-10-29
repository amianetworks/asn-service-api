# ASN Service API Design

> Interface contract between the ASN framework and service implementations.  
> Covers: call sequencing, re-entrancy, state machines, and implementation obligations.
>
> For per-method details, see inline comments in the respective package files.

---

## Open Items

Issues not yet resolved in the current API design. These must be decided before the affected surfaces are considered stable.

| # | Area | Item | Status |
|---|---|---|---|
| 1 | §5.1 Config op callbacks | Whether `ASNServiceController.AddConfigOps / UpdateConfigOp / DeleteConfigOps` are invoked at all is under discussion. The alternative: drop controller-side callbacks entirely and let config op errors propagate from the node directly to `Malfunctioning`, consistent with how `Start()` errors are handled. | Under discussion |

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Package Organization](#2-package-organization)
3. [Object Relationships](#3-object-relationships)
4. [State Management](#4-state-management)
5. [Controller Side — `capi`](#5-controller-side--capi)
6. [Service Node Side — `snapi`](#6-service-node-side--snapi)
7. [Config Ops](#7-config-ops)
8. [Ops Commands](#8-ops-commands)
9. [Cross-Service Data Sharing](#9-cross-service-data-sharing)
10. [IAM](#10-iam)
11. [Databases](#11-databases)
12. [Implementation Checklist](#12-implementation-checklist)

---

## 1. Architecture Overview

ASN is a distributed control-plane framework with a fixed topology: **one Controller** (management plane) paired with **N Service Nodes** (data plane). Services are loaded as **`.so` shared libraries** at runtime.

```mermaid
graph TD
    subgraph Framework["ASN Framework"]
        C["ASN Controller (single instance)<br/>impl: ASNController<br/>calls: ASNServiceController"]
        N["ASN Service Node ×N (one per node)<br/>impl: ASNServiceNode<br/>calls: ASNService"]
        C -->|manages N nodes| N
    end
```

The framework owns all topology (networks, nodes, groups); services observe and annotate it. Services do not communicate through the framework's network layer — intra-node cross-service data exchange uses the Shared Data mechanism (§9).

---

## 2. Package Organization

| Package | Path Suffix | Role |
|---|---|---|
| `capi` | `/controller` | Controller-side interfaces and structs |
| `snapi` | `/servicenode` | Service-node-side interfaces and structs |
| `commonapi` | `/common` | Shared enums, structs, DB/log abstractions |
| `iam` | `/iam` | IAM interface |
| `log` | `/log` | Structured logger interface |

---

## 3. Object Relationships

```mermaid
graph LR
    subgraph SVC["SERVICE-IMPLEMENTED"]
        SC[ASNServiceController]
        S[ASNService]
    end
    subgraph FW["FRAMEWORK-PROVIDED"]
        AC[ASNController]
        SN[ASNServiceNode]
    end
    AC -->|lifecycle| SC
    SC -->|calls| AC
    SN -->|lifecycle| S
    S -->|calls| SN
```

| Interface | Implemented by | Consumed by |
|---|---|---|
| `ASNController` | Framework | `ASNServiceController` |
| `ASNServiceController` | Service | Framework |
| `ASNServiceNode` | Framework | `ASNService` |
| `ASNService` | Service | Framework |
| `iam.Instance` | Framework | `ASNServiceController` via `GetIAM()` |

Both `ASNServiceController` and `ASNService` expose a `StaticResource` sub-interface callable **before** `Init()`.

---

## 4. State Management

### Node States (`commonapi.NodeState`)

Framework-owned; services cannot set it directly.

| Value | Constant | Meaning |
|---|---|---|
| `0` | `NodeStateUnregistered` | Never successfully registered |
| `1` | `NodeStateOffline` | Registered but unreachable |
| `2` | `NodeStateOnline` | Connected and reachable |
| `3` | `NodeStateMaintenance` | Online but in maintenance mode |

#### Node State Transition Diagram

```mermaid
stateDiagram-v2
    [*] --> Unregistered
    Unregistered --> Offline : register
    Offline --> Online : heartbeat
    Online --> Offline : disconnect
    Online --> Maintenance : enter maintenance
    Maintenance --> Online : exit maintenance
    Maintenance --> Offline : disconnect
```

First-time registration always lands in `Offline`. Subsequent heartbeats drive transitions between `Offline`, `Online`, and `Maintenance`. The framework broadcasts every state change via `SubscribeNodeStateChanges()`.

#### Per-State Capabilities

| | Unregistered | Offline | Online | Maintenance |
|---|---|---|---|---|
| Register node | ✓ | — | — | — |
| Subscribe to node stream | — | ✓ (required) | — | — |
| Send service commands (Start / Stop / Reset) | — | — | ✓ | ✓ |
| Send service ops (`SendServiceOps`) | — | ✗ `NodeDisconnected` | ✓ | ✓ |
| Send config ops | — | ✗ `NodeDisconnected` | ✓ | ✓ |
| Receive `NodeStateChange` events | — | ✓ | ✓ | ✓ |

**Side effect on going `Offline`:** The framework immediately sets every loaded service on that node to `ServiceStateUnavailable`. All in-flight ops return `FrameworkErrNodeDisconnected`.

**`SubscribeNodeStateChanges`** — one-shot; a second call errors. Delivers an initial snapshot of the service's nodes, then incremental changes. The channel is never closed during normal operation. Delivery is best effort: a change that finds the channel full is dropped, so a service reconciles at start and periodically.

Every `NodeStateChange` has a `Kind`; a consumer branches on it and ignores kinds it does not know.

| Kind | Fields | Meaning |
|---|---|---|
| `ChangeKindState` | `NodeState`, `FrameworkError`, `ServiceState`, `ServiceError` | Connectivity or this service's state changed (the only kind in the snapshot). |
| `ChangeKindNodeDeleted` | `NodeID` | The node was deleted. |
| `ChangeKindNodeServiceRemoved` | `NodeID` | This service left the node's record (before its deb is purged). |
| `ChangeKindNodeMoved` | `NodeID`, `FromNetworkID`, `NetworkID` | The node moved between networks. |
| `ChangeKindRootNetworkDeleted` | `NetworkID` | A root network was deleted (sent to every service). |
| `ChangeKindNodeTask` | `NodeID`, `TaskSlot`, `Task` | One of this service's task slots on the node was filled or cleared (`Task` nil). |
| `ChangeKindNodeAutoUpgrade` | `NodeID`, `AutoUpgrade` | The node's auto-upgrade switch changed (sent to every service on the node; no actor). |

Changes describe the record, not the host. The task slots and `Node.AutoUpgrade` are not in the snapshot; read them when subscribing.

---

### Service States (`commonapi.ServiceState`)

Tracked independently per node. The framework owns all transitions; the service influences them only through return values from `Init()`, `Start()`, `Stop()`, and config op methods.

| Value | Constant | Meaning |
|---|---|---|
| `0` | `ServiceStateUnavailable` | `.so` not loaded on this node |
| `1` | `ServiceStateUninitialized` | `.so` loaded; `Init()` not yet called or failed |
| `2` | `ServiceStateInitialized` | `Init()` succeeded; waiting for start |
| `3` | `ServiceStateConfiguring` | `Start()` in progress (transient) |
| `4` | `ServiceStateRunning` | `Start()` succeeded; fully operational |
| `5` | `ServiceStateMalfunctioning` | Fatal error; must be explicitly restarted |

#### Service State Transition Diagram

```mermaid
stateDiagram-v2
    direction TB
    [*] --> Unavailable

    Unavailable --> Uninitialized : load .so

    Uninitialized --> Initialized : Init ok
    Uninitialized --> Uninitialized : Init fail

    Initialized --> Unavailable : cluster registration fail
    Initialized --> Configuring : StartService
    Initialized --> Uninitialized : ResetService

    Malfunctioning --> Configuring : StartService

    Configuring --> Running : Start ok
    Configuring --> Malfunctioning : Start fail or timeout
    Configuring --> Initialized : StopService
    Configuring --> Configuring : ErrRestartNeeded

    Running --> Configuring : update config
    Running --> Initialized : StopService, or connection fail with auto-stop yes
    Running --> Malfunctioning : runtimeErrChan, ops fail, or connection fail with auto-stop no

    Malfunctioning --> Initialized : StopService or ResetService succeeds
    Malfunctioning --> Malfunctioning : StopService or ResetService fails
```

**`Init()` is triggered automatically** by the framework after `.so` load. `Uninitialized` is a stable state only when `Init()` fails — `Init()` failure does **not** produce `Malfunctioning`; the service stays retryable via `ResetService`.

**Cluster mode registration**: In cluster mode, a freshly `INITIALIZED` service registers with the controller before the first `Start()`. Registration failure returns the service to `UNAVAILABLE`.

**Auto-start**: Services may be configured to start automatically after initialization (and registration in cluster mode), or wait for an explicit `StartService` command.

**Connection fail behavior** depends on the service's `auto-stop` configuration:
- `auto-stop=yes` — service cleanly stops → `INITIALIZED` (ready to restart on reconnection)
- `auto-stop=no` — service stays running; errors eventually surface via `runtimeErrChan` → `Malfunctioning`

**Reset (full reload)**: Unloads and reloads the service including a fresh `Init()`. The service briefly passes through `Uninitialized` then auto-transitions to `Initialized`.

#### Per-State Capabilities

The table below answers: *when the service is in state X, which framework operations are allowed?*

| Operation | Unavailable | Uninitialized | Initialized | Configuring | Running | Malfunctioning |
|---|---|---|---|---|---|---|
| `CreateNode` (load + Init) | ✓ | — | — | — | — | — |
| `DeleteNode` (unload) | ✓ | ✓ | ✓ | ✓ (stops first) | ✓ (stops first) | ✓ (stops first) |
| `StartService` | — | — | ✓ | — | — | ✓ |
| `StopService` | — | — | — | ✓ | ✓ | ✓ |
| `ResetService` (full reload) | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| `StartService` with updated config (hot-reload) | — | — | — | ✓ | ✓ | ✓ |
| `SendServiceOps` / `ApplyServiceOps` | ✗ | ✗ | ✗ | ✗ | **✓ only** | ✗ |
| `AddConfigOps` | ✗ | ✗ | ✗ | ✗ | **✓ only** | ✗ |
| `UpdateConfigOp` | ✗ | ✗ | ✗ | ✗ | **✓ only** | ✗ |
| `DeleteConfigOps` | ✗ | ✗ | ✗ | ✗ | **✓ only** | ✗ |

`✗` returns `FrameworkErrServiceStateNotAllowed`. `—` means the operation is silently skipped or is inapplicable (not an error).

**Critical rule:** `ApplyServiceOps` and all three config op methods (`AddConfigOps`, `UpdateConfigOp`, `DeleteConfigOps`) are only dispatched to a service node when that service is in `Running`. Any other state produces `FrameworkErrServiceStateNotAllowed` in the `OpsResponse`.

#### State Transition Reference

| From | To | Trigger |
|---|---|---|
| `Unavailable` | `Uninitialized` | `.so` loaded (after `CreateNode` on an online node, or at registration); `Init()` auto-triggered |
| `Uninitialized` | `Initialized` | `Init()` returns `nil` |
| `Uninitialized` | `Uninitialized` | `Init()` returns error (stays; retryable via `ResetService`) |
| `Initialized` | `Unavailable` | Cluster mode: registration to controller fails |
| `Initialized` | `Configuring` | `StartService` dispatched (explicit or auto-start after init / registration) |
| `Malfunctioning` | `Configuring` | `StartService` dispatched |
| `Configuring` | `Running` | `Start()` returns `nil` |
| `Configuring` | `Malfunctioning` | `Start()` returns non-`ErrRestartNeeded` error, or start timeout |
| `Running` | `Configuring` | Update config command received; `Start(newConfig)` dispatched (hot-reload or after `Stop()`) |
| `Configuring` | `Configuring` | `Start()` returns `ErrRestartNeeded` → framework calls `Stop()` then `Start()` again (no re-`Init()`) |
| `Running` | `Malfunctioning` | `runtimeErrChan` receives a value |
| `Running` | `Malfunctioning` | Any config op method returns error |
| `Running` | `Malfunctioning` | `Stop()` times out or returns error |
| `Running` | `Malfunctioning` | Connection fail with `auto-stop=no` |
| `Running` | `Initialized` | Connection fail with `auto-stop=yes` (service cleanly stopped) |
| `Configuring` | `Initialized` | `StopService` succeeds |
| `Running` | `Initialized` | `StopService` succeeds |
| `Malfunctioning` | `Initialized` | `StopService` or `ResetService` succeeds |
| `Malfunctioning` | `Malfunctioning` | `StopService` or `ResetService` fails (stays) |
| `Initialized` | `Uninitialized` | `ResetService` (full reload): `Stop` → `Finish` → reload `.so` → `Init()` |
| Any | `Unavailable` | `DeleteNode` (unload) |

#### `ErrRestartNeeded` Semantics

When `Start()` returns `ErrRestartNeeded`, the framework interprets this as "hot-reload is not supported for this config delta." It automatically executes `Stop()` → `Start(newConfig)` — **without calling `Init()` again** — and the service stays in `Configuring` throughout. The net result is `Configuring → Running` or `Configuring → Malfunctioning` as usual; the intermediate stop is transparent to the caller.

---

### Config Source (`commonapi.ServiceSource`)

`Node.ServiceInfo.ConfigSource` — origin of a node's active config:

| Constant | Meaning |
|---|---|
| `ServiceConfigSourceNode` | Config set directly on the node |
| `ServiceConfigSourceNodeGroup` | Config inherited from the node's group |

---

## 5. Controller Side — `capi`

### 5.1 ASNServiceController

Lifecycle order (see `controller/service.go` for per-method contracts):

```
StaticResource()             any time; re-entrant
  ├─ Version()
  ├─ CLICommands()
  └─ WebHandler()

Init(asnController)          once; not re-entrant
Start(config)                after Init; sequential; repeatable

  ├─ HandleMessageFromNode() after Init; concurrent
  ├─ AddConfigOps()          after Init; concurrent       ⚠ see Open Item #1
  ├─ UpdateConfigOp()        after Init; concurrent       ⚠ see Open Item #1
  ├─ DeleteConfigOps()       after Init; concurrent       ⚠ see Open Item #1
  └─ GetMetrics()            after Init; concurrent

Stop()                       idempotent
Finish()                     once; after Stop
```

### 5.2 ASNController

All methods goroutine-safe after `Init()`. See `controller/asn.go` for the full API.

**Node tasks.** `CreateNode` and `DeleteNode` change the host themselves on a node that can run node tasks: the result carries `TaskID`, or `TaskPending` with `PendingReason`, and `Token` is nil. Any other node gets a token as before; the choice depends only on the node, never on a failure. `UpgradeNode` upgrades `asnsn` and the calling service's deb. Every package task restarts `asnsn` and all services on the node.

**Task status.** `CreateNode` and `UpgradeNode` return a `RequestID`. The service follows it through `ServiceInfo.PendingTask` (queued), `RunningTask` (sent) and `LastRequest` (ended); the request is over when `LastRequest.RequestID` equals it. `LastTask` holds the last task of any actor that restarted the node; another actor's task carries no `RequestID` and shows as `upgrade`, `install`, or `restart`, naming no actor, service, or package. Status is `OK` or `Failed`; the cause is for an ASN operator.

**Auto-upgrade.** `SetNodeAutoUpgrade` sets a node-wide switch that any service on the node, or an ASN operator, may change. With it on, the controller upgrades the whole node to the builds it serves, as long as the node is inside the controller's version window, one attempt per target version set (`Node.AutoUpgradeHeld` after a failure); every service on the node restarts. A node outside the window (in Maintenance) is never upgraded automatically.

**Root networks.** A service may create, rename, and delete root networks (for example one per tenant), change the members of any root's admin group (one per root, created and deleted with the root by ASN, and shared by every service and by ASN operators), move its nodes between roots, and list networks with a filter. These methods act on any root network, the default root included, except that a service can neither delete nor rename the default root. Deleting a root deletes every node in it, of every service; hosts are not cleaned up. A root given the configured default-root name may become the default root; check `Network.DefaultRoot` after creating one.

---

## 6. Service Node Side — `snapi`

### 6.1 ASNService

Lifecycle order (see `servicenode/service.go` for per-method contracts):

```
StaticResource()              any time; re-entrant
  ├─ Version()
  └─ SharedData()

Init(asnServiceNode)          once; not re-entrant
Start(config)                 after Init; sequential; repeatable → runtimeErrChan

  ├─ ApplyServiceOps()        after Start; explicitly concurrent
  ├─ AddConfigOps()           after Start; concurrent
  ├─ UpdateConfigOp()         after Start; concurrent
  ├─ DeleteConfigOps()        after Start; concurrent
  ├─ OnQuerySharedData()      after Init; concurrent
  └─ OnSubscribeSharedData()  after Init; concurrent

Stop()                        idempotent
Finish()                      once; after Stop
```

### 6.2 ASNServiceNode

All methods goroutine-safe after `Init()`. See `servicenode/asn.go` for the full API.

---

## 7. Config Ops

Persistent, incremental configuration directives stored per node or node group. Complement the monolithic YAML config passed to `Start()`.

### Data Flow

```mermaid
sequenceDiagram
    participant C as Controller
    participant F as Framework
    participant SN as Service Node

    C->>F: ASNController.AddConfigOps()
    F->>F: persist
    F->>SN: ASNService.AddConfigOps()
    SN-->>F: OpsResponse
    F-->>C: OpsResponse chan
    F->>C: ASNServiceController.AddConfigOps() ⚠ Open Item #1
```

Framework persists and dispatches to nodes **before** invoking `ASNServiceController.AddConfigOps()`.

### Scoping Rules

- Scope: `ServiceScopeNodeGroup`(3) or `ServiceScopeNode`(4) only.
- Group-level ops propagate to all member nodes unless a node has direct overrides.
- `ListConfigOps` returns ops directly on the specified scope; does not traverse group→node hierarchy.
- `ConfigOp.ID` is framework-assigned; use it for `UpdateConfigOp` and `DeleteConfigOps`.
- `ConfigOp.ConfigParams` is an opaque service-defined string.

---

## 8. Ops Commands

Ephemeral, on-demand directives from controller to service nodes. Not persisted.

| Method | Dispatch | Blocking | Scope |
|---|---|---|---|
| `SendServiceOps()` | Fan-out | No | Multiple nodes via `ServiceScope` |
| `SendServiceOpsToNode()` | Point-to-point | Yes | Single node by ID |

`opCmd` and `opParams` are service-defined. Controller and service node must share per-command schemas (typically JSON-encoded structs). See `ASNController.SendServiceOps` in `controller/asn.go` for usage example.

---

## 9. Cross-Service Data Sharing

Services on the same node exchange data through the framework's local registry without network overhead.

| Model | Consumer API | Provider callback | Use case |
|---|---|---|---|
| Query (pull) | `QueryServiceSharedData` | `OnQuerySharedData` | Point-in-time snapshot |
| Subscribe (push) | `SubscribeServiceSharedData` | `OnSubscribeSharedData` | Continuous / time-series stream |

Key names and value types are service-defined and opaque to the framework — producing and consuming services must agree out-of-band. Only one active subscription per key is permitted at a time; re-subscribe only after the channel closes.

---

## 10. IAM

Obtained via `ASNController.GetIAM()`. Full API: `iam/api.go`.

### Authentication Flow

All `LoginOrCreate*` and `AccountPasskeyAuth` return `(account, needMfa, tokenSet, err)`.  
When `needMfa == true`, `tokenSet` is a pre-MFA token; call `MFALoginVerify()` to fully authorize the session.

### Multi-Step Flows

| Flow | Steps |
|---|---|
| TOTP enrollment | `TotpBind` → `TotpBindConfirm` |
| Passkey registration | `AccountPasskeyBindChallengeGet` → `AccountPasskeyBind` |
| Passkey login | `AccountPasskeyLoginChallengeGet` → `AccountPasskeyAuth` |
| Phone/Email OTP | `AccountPhoneSend` / `AccountEmailSend` → `LoginOrCreate*` or `MFALoginVerify` |

### Group → Access Model

```
Group ──members──► Account(s)
  └──accesses──► Access(name, scope, operation, TimeControl)
```

Accesses are granted at the group level. `AccountAccessList` returns all effective rules for an account derived from its group memberships.

### Network Scope

`Account.ManagedNetworks` lists the ASN networks an account may act on, from every network-type group it belongs to, computed live (never copied or cached). Every service receives the same value and grants by it only on resources it places under ASN networks; a scope covers the subtree; each service decides which operations it honors; `ServiceAdmin` keeps its full access and does not extend to other services. Manage on a root network itself comes only from that root's admin group, so its members are the root's complete list of administrators; any other network-type group holding a root grants view only, and no network-type group spans two roots. `NetworkScope.RootID` names the root of each scope. Any service may add and remove members of any root's admin group through `RootAdminGroup*Members`, one change at a time and never by replacing the list with one it keeps; added members must already be in the calling service's namespace. Service admin groups are kept by ASN and are read-only to services.

---

## 11. Databases

`InitDocDB` / `InitTSDB` return `commonapi.DocDBHandler` / `commonapi.TSDBHandler`. The API
does not depend on am.modules: records are `map[string]any`, and time-series points and
queries are the plain structs `TSPoint` / `TSQuery`.

Doc records are selected with `commonapi.DocQuery`, built with `DocEq`, `DocIn`, `DocGt`,
`DocRegex`, `DocAnd` and the other `Doc*` functions. It is plain data, field for field the
query of am.modules, which validates it when a collection runs it; the framework copies it
without looking at the operators, so an operator am.modules adds needs no change here
beyond its constant and builder. The zero query is an error: `DocAll()` selects every
record, and `DocAllOf(conds...)` every record when `conds` is empty. A regular expression is
case-sensitive unless its options hold `i`. `FindRecords` takes options to sort, page and
project the records (`DocAsc`, `DocDesc`, `DocSkip`, `DocLimit`, `DocPage`, `DocFields`,
`DocOmit`), the same as the find options of am.modules, which validates them; `FindRecord`
takes `DocFields` and `DocOmit` only. The framework reads them with `DocResolveFindOptions`.

Times in records are stored as RFC 3339 strings, which do not compare in time order, so a
record to be selected or sorted by time also keeps it as a number, e.g. Unix microseconds,
and the query compares that number.

A service that wants the typed am.modules API (`docdb.Collection[T]`, `*tsdb.Point`,
`*tsdb.Query`), or an am.modules module built on it such as subscription, wraps the handlers
in its own adapter to `docdb.Database` / `tsdb.Database`. Such adapters live in the service,
never in this repository. In-App Subscription is no longer provided by the framework; a
service uses the am.modules subscription module directly with the adapted DocDB and its own
redis client from `InitRedis`.

---

## 12. Implementation Checklist

### ASNServiceController

- [ ] `StaticResource()` and sub-methods: pre-`Init()` safe, re-entrant
- [ ] `CLICommands()`: purely declarative, no side effects
- [ ] `WebHandler()`: no resource acquisition
- [ ] `Init()`: calls each `Init*` / `Get*` exactly once; no goroutines
- [ ] `Start()`: returns promptly; fully supersedes prior config
- [ ] `HandleMessageFromNode()`: guards shared state (concurrent)
- [ ] Config op callbacks: guard shared state (concurrent)
- [ ] `GetMetrics()`: returns from pre-computed snapshots
- [ ] `Stop()`: idempotent, returns promptly
- [ ] `Finish()`: releases all resources; no goroutines remain

### ASNService

- [ ] `SharedData()`: accurately declares all provided keys, or `(nil, nil)`
- [ ] `Init()`: calls each `Init*` exactly once; no goroutines
- [ ] `Start()`: idempotent; returns `ErrRestartNeeded` if hot-reload unsupported
- [ ] `runtimeErrChan`: only for unrecoverable post-start failures
- [ ] `ApplyServiceOps()`: synchronizes all shared state (explicitly concurrent)
- [ ] Config op callbacks: error only for unrecoverable failures
- [ ] `OnQuerySharedData()`: returns `ErrKeyNotFound` for undeclared keys
- [ ] `OnSubscribeSharedData()`: closes every returned channel after stream ends
- [ ] `Stop()`: idempotent, returns promptly
- [ ] `Finish()`: releases all resources; no goroutines remain

### Cross-Cutting

- [ ] `SubscribeNodeStateChanges()`: called at most once; every change handled by `Kind`, unknown kinds ignored; state reconciled at start and periodically
- [ ] `CreateNode` / `DeleteNode`: a nil `Token` with a `TaskID` or `TaskPending` is handled, and a result with none of them means done
- [ ] `CreateNode` / `UpgradeNode`: a request is followed by `RequestID` until `ServiceInfo.LastRequest` carries it; no own timeout
- [ ] `SetNodeAutoUpgrade`: called on a user's decision, not re-asserted periodically; the user is told that every service on the node restarts
- [ ] `ManagedNetworks`: honored only on network-placed resources
- [ ] `ManagedNetworks`: a scope on a root is recognized by `RootID == NetworkID`; scopes are grouped by `RootID`
- [ ] Root admin group: members added and removed one at a time, never reconciled against a list the service keeps
- [ ] `runtimeErrChan`: carries only fatal errors, not service-internal errors
- [ ] `opCmd` / `opParams` schemas: defined and shared across controller and service node
- [ ] Shared data keys and value types: agreed upon out-of-band between service teams
- [ ] Config op payload format: versioned for rolling upgrade compatibility
