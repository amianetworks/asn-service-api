# ASN Service API Design

> Interface contract between the ASN framework and service implementations.  
> Covers: call sequencing, re-entrancy, state machines, and implementation obligations.

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

ASN is a distributed control-plane framework with a fixed topology: **one Controller** (management plane) paired with **N Service Nodes** (data plane). Services are loaded as **`.so` shared libraries** at runtime.

```
┌────────────────────────────────────────────────────────┐
│                     ASN Framework                      │
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

The framework owns all topology (networks, nodes, groups); services observe and annotate it. Services do not communicate through the framework's network layer — intra-node cross-service data exchange uses the Shared Data mechanism (§9).

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
  ┌─────────────────────┐               ┌─────────────────────┐
  │ ASNServiceController│◄─ lifecycle ──│   ASNController     │
  │                     │──── calls ───►│                     │
  └─────────────────────┘               └─────────────────────┘
  ┌─────────────────────┐               ┌─────────────────────┐
  │     ASNService      │◄─ lifecycle ──│   ASNServiceNode    │
  │                     │──── calls ───►│                     │
  └─────────────────────┘               └─────────────────────┘
```

| Interface | Implemented by | Consumed by |
|---|---|---|
| `ASNController` | Framework | `ASNServiceController` |
| `ASNServiceController` | Service | Framework |
| `ASNServiceNode` | Framework | `ASNService` |
| `ASNService` | Service | Framework |
| `iam.Instance` | Framework | `ASNServiceController` via `GetIAM()` |
| `subscription.Instance` | Framework | `ASNServiceController` via `GetSubscription()` |

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

```
Unregistered ──► Offline ◄──► Online ◄──► Maintenance
```

`NodeStateChange` events carry the updated `NodeState`, current `ServiceState`, `FrameworkError` (non-nil on framework-level failures), and `ServiceError` (non-nil when the service reported an error during transition). In-flight ops to an `Offline` node yield `FrameworkErrNodeDisconnected`.

### Service States (`commonapi.ServiceState`)

Tracked independently per node.

| Value | Constant | Meaning |
|---|---|---|
| `0` | `ServiceStateUnavailable` | `.so` not loaded |
| `1` | `ServiceStateUninitialized` | Loaded; `Init()` not yet succeeded |
| `2` | `ServiceStateInitialized` | `Init()` succeeded |
| `3` | `ServiceStateConfiguring` | Applying config or config ops |
| `4` | `ServiceStateRunning` | `Start()` succeeded |
| `5` | `ServiceStateMalfunctioning` | Fatal error or config op failure |

```
Unavailable ──(AddServiceToNode)──► Uninitialized ──(Init ok)──► Initialized
                                                                      │
                                    ◄──(StopService ok)───────────────┤
                                                                      │(Start)
                                                               Configuring ◄──► Running
                                                                                   │
                                                                             Malfunctioning
```

**Service-side state triggers:**

| Trigger | Effect |
|---|---|
| `Init()` error | Stays `Uninitialized` |
| `Start()` → `ErrRestartNeeded` | Framework: `Stop()` → `Start()` with new config (no re-init) |
| `Start()` → other error | → `Malfunctioning` |
| `runtimeErrChan` receives value | → `Malfunctioning` |
| Config op method returns error | → `Malfunctioning` |

**`FrameworkError` values in `OpsResponse`:**

| Error | Cause |
|---|---|
| `FrameworkErrServiceTimeout` | Node reachable; service did not respond in time |
| `FrameworkErrNodeDisconnected` | Node offline |
| `FrameworkErrServiceUnavailable` | Service not loaded on node |
| `FrameworkErrServiceStateNotAllowed` | Service not in `Running` state |

When `FrameworkError != nil`, `ServiceResponse` and `ServiceError` are undefined.

### Config Source (`commonapi.ServiceSource`)

`Node.ServiceInfo.ConfigSource` — origin of a node's active config:

| Constant | Meaning |
|---|---|
| `ServiceConfigSourceNode` | Config set directly on the node |
| `ServiceConfigSourceNodeGroup` | Config inherited from the node's group |

---

## 5. Controller Side — `capi`

### 5.1 ASNServiceController (service-implemented)

#### Lifecycle

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

#### Method Contracts

**`StaticResource()`** — re-entrant; returns a stable immutable object.

**`StaticResource.Version()`** — hardcoded version constant.

**`StaticResource.CLICommands(applyCLIOps)`** — returns Cobra commands for the ASN CLI. Purely declarative; no side effects. `applyCLIOps` is the framework dispatcher with identical semantics to `SendServiceOps`.

**`StaticResource.WebHandler(staticPath)`** — returns the `http.Handler` mounted by the framework under a service-specific prefix. `staticPath` is the static asset directory. No resource acquisition here.

**`Init(asnController)`** — called once before `Start()`. All `Init*` / `Get*` calls on `asnController` must happen here (one-shot). CLI commands must be runnable on success. No background goroutines. Failure → `ServiceStateUninitialized`.

**`Start(config)`** — sequential, repeatable. Must return promptly. Each call fully supersedes the previous config.

**`AddConfigOps / UpdateConfigOp / DeleteConfigOps`** — called after the framework has persisted ops and dispatched to nodes. Implement controller-side bookkeeping (routing, policy, in-memory state). Concurrent with other callbacks; guard shared state. `serviceScope` ∈ {`ServiceScopeNodeGroup`(3), `ServiceScopeNode`(4)}.

**`HandleMessageFromNode(nodeID, messageType, payload)`** — concurrent; messages from multiple nodes may arrive simultaneously. No direct response channel; to reply, use `SendServiceOpsToNode()`.

**`GetMetrics(networkID)`** — concurrent; must return promptly from pre-computed snapshots. Values must be JSON-serializable.

**`Stop()`** — idempotent; must return promptly. Return value is informational; framework shuts down regardless.

**`Finish()`** — called once before `.so` unload. Goroutines remaining after return and touching service memory cause undefined behavior.

---

### 5.2 ASNController (framework-provided)

All methods goroutine-safe after `Init()`.

#### Resource Initialization

Call **in `Init()`**. One-shot per resource name; second call returns error.

| Method | Constraint | Returns |
|---|---|---|
| `InitLogger()` | Once | `*log.Logger` (R/A/P/E); see §12.4 |
| `InitDocDB(name)` | Once per name | Connected `DocDBHandler` |
| `InitTSDB(name)` | Once per name | Connected `TSDBHandler` |
| `InitLocker()` | Once | Cluster-wide `Lock` |
| `GetIAM()` | Once | `iam.Instance`; see §10 |
| `GetSubscription()` | Once | `subscription.Instance`; see §11 |

#### Service Lifecycle

| Method | Effect |
|---|---|
| `AddServiceToNode(nodeID)` | Loads `.so` onto node, triggers node `Init()`. Node must be online. |
| `DeleteServiceFromNode(nodeID)` | Calls `Stop()` + `Finish()` then unloads `.so`. Not a substitute for `StopService`. |
| `StartService(scope, list)` | Triggers `Start(config)` on matched nodes |
| `StopService(scope, list)` | Triggers `Stop()` on matched nodes |
| `ResetService(scope, list)` | `Stop()` then `Start()` on matched nodes |

**`ServiceScope` / `scopeList` mapping:**

| Constant | Value | `scopeList` |
|---|---|---|
| `ServiceScopeNetwork` | 1 | Network IDs |
| `ServiceScopeNetworkWithSubnetworks` | 2 | Network IDs (recursive) |
| `ServiceScopeNodeGroup` | 3 | Node Group IDs |
| `ServiceScopeNode` | 4 | Node IDs |

#### Ops Dispatch

**`SendServiceOps(scope, list, opCmd, opParams) (<-chan *OpsResponse, error)`** — fan-out, async. `error != nil` → invalid scope/list, channel nil. Otherwise returns immediately; channel closed after all nodes respond or time out.

```go
resChan, paramErr := ctrl.SendServiceOps(scope, list, cmd, params)
if paramErr != nil { ... }
for res := range resChan {
    if res.FrameworkError != nil { ... }
}
```

**`SendServiceOpsToNode(nodeID, opCmd, opParams) (*OpsResponse, error)`** — point-to-point, synchronous. `error != nil` → invalid `nodeID`.

#### Config Ops Dispatch

Scope limited to `ServiceScopeNodeGroup`(3) or `ServiceScopeNode`(4).

| Method | Effect |
|---|---|
| `AddConfigOps(scope, scopeID, configParams)` | Persists ops, fans out to nodes; returns response channel |
| `UpdateConfigOp(scope, scopeID, configOpID, configParam)` | Updates one op by ID, fans out |
| `DeleteConfigOps(scope, scopeID, configOpIDs)` | Removes ops by ID, fans out |
| `ListConfigOps(scope, scopeID)` | Returns ops directly on scope; synchronous, no fan-out |

Return semantics for fan-out methods match `SendServiceOps`: `error != nil` → invalid scope/ID; otherwise channel closed after all responses.

`ListConfigOps` does **not** traverse the group-to-node inheritance hierarchy.

#### Node Topology

| Method | Returns |
|---|---|
| `GetNetworks()` | Full network tree; each `Network` embeds child `Networks` |
| `GetNodeByID(nodeID)` | Hardware info + service `Metadata` + `ServiceInfo` (state, config source, ops) |
| `UpdateNodeMetadata(nodeID, metadata)` | Persists opaque string; retrievable via `GetNodeByID().Metadata` |
| `SetConfigOfNode(nodeID, config)` | Persists YAML/UTF-8 config; used on next `StartService` |
| `GetNodesOfNetwork(networkID, withService)` | Nodes + links; `withService=true` filters to nodes with service loaded |

`GetNodesOfNetwork` link types:
- **Internal**: both endpoints in network; `To` node included in result.
- **External**: `To` outside network; not included in result.

**`SubscribeNodeStateChanges() (<-chan *NodeStateChange, error)`** — one-shot (second call errors). Delivers initial snapshot of all nodes, then incremental changes. Channel never closed during normal operation.

#### Node Group Management

All re-entrant.

| Method | Effect |
|---|---|
| `CreateNodeGroup(networkID, name, desc, metadata)` | Creates group scoped to this service |
| `ListNodeGroups(networkID)` | Lists this service's groups in the network |
| `GetNodeGroupByID(nodeGroupID)` | Returns group with `Metadata` and active `ConfigOps` |
| `UpdateNodeGroupMetadata(nodeGroupID, metadata)` | Persists service metadata on group |
| `DeleteNodeGroup(nodeGroupID)` | Deletes group; member nodes unaffected |
| `SetConfigOfNodeGroup(nodeGroupID, config)` | Persists config; inherited by members unless overridden |
| `AddNodesToNodeGroup(nodeGroupID, nodeIDs)` | Adds nodes to group |
| `RemoveNodesFromNodeGroup(nodeGroupID, nodeIDs)` | Removes nodes from group |

---

## 6. Service Node Side — `snapi`

### 6.1 ASNService (service-implemented)

#### Lifecycle

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

#### Method Contracts

**`StaticResource()`** — re-entrant; stable object.

**`StaticResource.Version()`** — hardcoded constant.

**`StaticResource.SharedData() (aggregated, subscribable []string)`** — declares provided shared data keys. `aggregated`: pull-model. `subscribable`: push/stream model. Return `(nil, nil)` if not a provider. Framework uses this to validate consumer requests.

**`Init(asnServiceNode)`** — once, not re-entrant. All `Init*` calls on `asnServiceNode` here. No goroutines. Failure → `ServiceStateUninitialized`.

**`Start(config) (runtimeErrChan, err)`** — sequential, repeatable. Must be idempotent; each call fully supersedes previous config. Return `ErrRestartNeeded` if hot-reload unsupported → framework executes `Stop()` → `Start()` (no re-init). `runtimeErrChan`: for post-start fatal errors only → `Malfunctioning` on receipt. Synchronous non-`ErrRestartNeeded` error → `Malfunctioning`.

**`ApplyServiceOps(opCmd, opParams) (resp, err)`** — **explicitly concurrent**; synchronize all shared state. `resp` → `OpsResponse.ServiceResponse`; `err` → `OpsResponse.ServiceError`. Neither is fatal; use `runtimeErrChan` for unrecoverable failures. Framework enforces per-call timeout; return promptly.

**`AddConfigOps / UpdateConfigOp / DeleteConfigOps`** — concurrent; guard shared state. Any returned error → `Malfunctioning`. Only return errors for unrecoverable failures.

**`OnQuerySharedData(keys) (map[string]any, error)`** — concurrent. Implement only if `aggregated` keys declared. Return `ErrKeyNotFound` for undeclared keys.

**`OnSubscribeSharedData(keys) (map[string]<-chan any, error)`** — concurrent. Implement only if `subscribable` keys declared. **Each channel must be closed** after all values sent. Return `ErrKeyNotFound` for undeclared keys.

**`Stop()`** — idempotent; return promptly. Any error may influence service state.

**`Finish()`** — once, before `.so` unload. Same constraint as controller-side: no goroutines after return.

---

### 6.2 ASNServiceNode (framework-provided)

All methods goroutine-safe after `Init()`.

#### Resource Initialization

One-shot per name; call in `Init()`.

| Method | Constraint | Returns |
|---|---|---|
| `InitLogger()` | Once | `*log.Logger` |
| `InitDocDB(name)` | Once per name | Connected `DocDBHandler` |
| `InitTSDB(name)` | Once per name | Connected `TSDBHandler` |

#### Node Information

**`GetNodeType()`** — returns `commonapi.NodeType`; see §13.

**`GetNodeInfo()`** — returns `snapi.NodeInfo` embedding `commonapi.NodeInfo` (interfaces, IPMI, management, hardware) plus node `ID` and active `ConfigOps` string list.

#### Communication

**`SendMessageToController(messageType, payload)`** — fire-and-forget upcall to `HandleMessageFromNode()`. No response. To receive a reply, the controller initiates `SendServiceOpsToNode()`.

#### Cross-Service Data Access

**`GetSharedData(serviceName)`** — returns key sets of another service on this node. `ErrServiceNotFound` if not loaded.

**`QueryServiceSharedData(serviceName, keys)`** — pull-model fetch. `ErrServiceNotFound` or `ErrKeyNotFound` on failure.

**`SubscribeServiceSharedData(serviceName, keys)`** — push-model; one channel per key, closed by provider at end of stream. **One active subscription per key only**; second subscription before first closes is undefined behavior. `ErrServiceNotFound` or `ErrKeyNotFound` on failure.

---

## 7. Config Ops

Persistent, incremental configuration directives stored per node or node group. Complement the monolithic YAML config passed to `Start()`.

### Data Flow

```
Controller                    Framework                 Service Node
ASNController.AddConfigOps()
  ├─ persist ────────────────────────────────────────► ASNService.AddConfigOps()
  └─ OpsResponse chan ◄────────────────────────────────
ASNServiceController.AddConfigOps()  ← controller-side hook (⚠ Open Item #1)
```

Framework persists and dispatches to nodes **before** invoking `ASNServiceController.AddConfigOps()`.

### Scoping Rules

- Scope: `ServiceScopeNodeGroup`(3) or `ServiceScopeNode`(4) only.
- Group-level ops propagate to all member nodes unless a node has direct overrides.
- `ListConfigOps` returns ops directly on the specified scope; does not traverse group→node hierarchy.
- `ConfigOp.ID` is framework-assigned; use it for update and delete.
- `ConfigOp.ConfigParams` is an opaque service-defined string.

### Error Semantics

A non-nil error from any `ASNService` config op method immediately → `Malfunctioning`. Return errors only for unrecoverable conditions.

---

## 8. Ops Commands

Ephemeral, on-demand directives from controller to service nodes. Not persisted.

| Method | Dispatch | Blocking | Scope |
|---|---|---|---|
| `SendServiceOps()` | Fan-out | No | Multiple nodes via `ServiceScope` |
| `SendServiceOpsToNode()` | Point-to-point | Yes | Single node by ID |

`opCmd` and `opParams` are service-defined. Controller and service node must share per-command schemas (typically JSON-encoded structs).

```go
// Controller
params, _ := json.Marshal(MyOpParams{...})
resChan, _ := ctrl.SendServiceOps(scope, list, "cmd", string(params))

// Service node — ApplyServiceOps
var p MyOpParams
json.Unmarshal([]byte(opParams), &p)
```

**`OpsResponse` fields:**

| Field | Content |
|---|---|
| `NodeID` | Responding node |
| `Timestamp` | Response time |
| `FrameworkError` | Non-nil if framework could not reach/invoke the service |
| `ServiceResponse` | `resp` from `ApplyServiceOps()`; valid only when `FrameworkError == nil` |
| `ServiceError` | `err` from `ApplyServiceOps()`; valid only when `FrameworkError == nil` |

---

## 9. Cross-Service Data Sharing

Services on the same node exchange data through the framework's local registry without network overhead.

| Model | Consumer API | Provider callback | Use case |
|---|---|---|---|
| Query (pull) | `QueryServiceSharedData` | `OnQuerySharedData` | Point-in-time snapshot |
| Subscribe (push) | `SubscribeServiceSharedData` | `OnSubscribeSharedData` | Continuous / time-series stream |

**Provider obligations:**
- Declare keys in `StaticResource.SharedData()` before `Init()`.
- Implement callbacks for declared key sets.
- Close each subscription channel after all values are sent.
- Return `ErrKeyNotFound` for undeclared keys.

**Consumer obligations:**
- Call `GetSharedData(serviceName)` to discover available keys.
- One active subscription per key; re-subscribe only after channel closes.

Key names and value types are service-defined and opaque to the framework; producing and consuming services must agree out-of-band.

---

## 10. IAM

Obtained via `ASNController.GetIAM()`. `iam.Instance` is goroutine-safe.


### Account

| Field | Type | Notes |
|---|---|---|
| `ID` | `string` | UUID |
| `Username` | `string` | Login username |
| `UsernameModified` | `bool` | True if username was changed from initial value |
| `Metadata` | `string` | Service-defined opaque string |
| `DeviceLimit` | `int` | Max concurrent logged-in devices |
| `Password` | `bool` | Password credential set |
| `Phone` | `Phone{CountryCode, Number}` | Bound phone |
| `Email` | `string` | Bound email |
| `Totp` | `bool` | TOTP enrolled |
| `MfaEnabled` | `bool` | MFA required for this account |
| `WeChat` | `*AccountWeChatInfo` | `{Bound, Nickname, HeadImgURL}` |
| `Apple` | `*AccountAppleInfo` | `{Bound, Email, EmailVerified, IsPrivateEmail}` |
| `Google` | `*AccountGoogleInfo` | `{Bound, Email, EmailVerified, Name, Picture}` |
| `Passkeys` | `[]AccountPasskey` | `{ID, DeviceID, CredentialID}` |
| `ServiceAdmin` | `bool` | Managed by ASN Controller, not by the service. The service cannot create, delete, or modify these accounts, but should grant them full access. Appears in account list results. |
| `Groups` | `[]string` | Group memberships |
| `Devices` | `map[string]*Device` | Active devices by device ID |
| `TimeInfo` | `TimeInfo{CreatedAt, UpdatedAt}` | Timestamps |

### Account Management

| Category | Methods |
|---|---|
| CRUD | `AccountCreate`, `AccountDelete`, `AccountGet`, `AccountList`, `AccountListByIDs`, `AccountExists` |
| Profile | `AccountRename`, `AccountMetadataUpdate`, `AccountPhoneUpdate`, `AccountEmailUpdate` |
| Third-party | `AccountWeChatUpdate`, `AccountAppleUpdate`, `AccountGoogleUpdate` |
| Password | `AccountPasswordUpdate` (requires old password), `AccountPasswordReset` (admin, no old password) |

`AccountCreate` — empty/nil credential fields are optional; `skipEmailValidation` / `skipPhoneValidation` bypass OTP for admin-initiated creation.

`AccountGet` — any single non-empty field serves as the lookup key.

`AccountList` — empty strings match all values for that field.

### OTP Sending

```go
AccountPhoneSend(countryCode, number string, sendWhenExistOnly bool) (string, error)
AccountEmailSend(email string, sendWhenExistOnly bool) (string, error)
```

`sendWhenExistOnly=true`: silently no-ops if no account exists for the contact.

### Authentication

`LoginMethods()` returns which methods are enabled at runtime.

| Method | Entry point |
|---|---|
| Username/email/phone + password | `LoginOrCreateWithPassword` |
| Phone OTP | `LoginOrCreateWithPhone` |
| Email OTP | `LoginOrCreateWithEmail` |
| WeChat OAuth | `LoginOrCreateWithWeChat` |
| Apple Sign-In | `LoginOrCreateWithApple` |
| Google Sign-In | `LoginOrCreateWithGoogle` |
| Passkey (WebAuthn) | `AccountPasskeyLoginChallengeGet` → `AccountPasskeyAuth` |

All `LoginOrCreate*` and `AccountPasskeyAuth` return `(account, needMfa, tokenSet, err)`. When `needMfa==true`, `tokenSet` is pre-MFA; session not fully authorized until `MFALoginVerify` succeeds. Common parameters: `device` (required), `userClaims` (embedded in token), `durationAccess/durationRefresh`, `createIfNotExist`.

`PasswordVerify` — validates credentials without issuing tokens.

`Logout(accountID, deviceID)` — invalidates session and revokes tokens for that device.

`AppleRedirect(w, r)` — Apple OAuth redirect handler; mount at the Apple-configured redirect URI.

### Token Lifecycle

| Method | Behavior |
|---|---|
| `TokenVerify(token)` | Returns `(mfaNeeded, accountID, username, deviceID, userClaims, err)` |
| `TokenRefresh(userClaims, tokenSet, duration)` | New access token from valid refresh token |
| `TokenRevoke(token)` | Immediately invalidates access token |
| `JWKSGet()` | JWKS document for external token verification |

`mfaNeeded==true` → pre-MFA token; only MFA endpoints should accept it.

### MFA

`AccountEnableMFA(accountID)` / `AccountDisableMFA(accountID)` — per-account control.

`ServiceMfaSet(bool)` / `MfaEnforced()` — service-wide enforcement.

| Type | Constant | Flow |
|---|---|---|
| TOTP | `MfaTypeTotp` | `TotpBind` → `TotpBindConfirm`; verify: `MFALoginVerify` |
| Email OTP | `MfaTypeEmail` | `AccountEmailSend` → `MFALoginVerify` |
| Phone OTP | `MfaTypeSms` | `AccountPhoneSend` → `MFALoginVerify` |
| Passkey | `MfaTypePasskey` | `AccountPasskeyLoginChallengeGet` → `MFALoginVerify` |

`TotpBind(accountID)` — returns QR image (data URI), issuer, and raw secret.

`MFALoginVerify(token, method, code, domain, sessionID, data)` — upgrades pre-MFA token. Unused method fields should be empty strings.

### Passkey (WebAuthn)

```
Bind:   AccountPasskeyBindChallengeGet(domain, accountID) → sessionID, data
        AccountPasskeyBind(domain, accountID, sessionID, deviceID, data)

Login:  AccountPasskeyLoginChallengeGet(domain) → sessionID, data
        AccountPasskeyAuth(device, userClaims, durationAccess, durationRefresh, domain, sessionID, data)

Unbind: AccountPasskeyUnbind(accountID, passkeyID)
        AccountPasskeyUnbindAll(accountID)
```

### Device Management

Login calls require `*DeviceInfo`. Framework tracks one `Device` per (account, device) pair; `Device.ID` is UUID assigned on first login.

| Field | Type | Values |
|---|---|---|
| `Category` | `DeviceCategory` | `Phone`, `Tablet`, `Wearable`, `Browser`, `MiniProgram`, `PC` |
| `Type` | `DeviceType` | Per-category (e.g., `PhoneIPhone`, `PhoneAndroid`, `PCMacOS`) |
| `OS` | `DeviceOs` | `ios`, `android`, `watchos`, `android_wear`, `wechat`, `windows`, `macos`, `linux` |
| `Language` | `DeviceLanguage` | `EN`, `ZH` |
| `Name/Model/SerialNumber/PushToken/Metadata` | `string` | — |

| Method | Effect |
|---|---|
| `DeviceLimitUpdate(accountID, limit)` | Sets max concurrent devices |
| `DeviceInfoUpdate(accountID, device)` | Updates stored device record |
| `DeviceDelete(accountID, deviceID)` | Removes device; revokes its tokens |

### Groups

```
Group ──members──► Account(s)
  └──accesses──► Access(name, scope, operation, TimeControl)
```

| Method | Effect |
|---|---|
| `GroupCreate/Delete/Exists/Rename/MetadataUpdate` | Basic CRUD |
| `GroupGet(name)` | Returns `*Group{GroupName, GroupMembers, Metadata, TimeInfo}` |
| `GroupList()` | Lists all groups in this service's namespace |
| `GroupMemberList(name)` | Returns member `*Account` list |
| `AccountJoinGroup/AccountLeaveGroup` | Membership management |
| `AccountGroupList(accountID)` | Groups this account belongs to |

### Accesses

An Access is a named, time-controlled permission rule.

| Method | Effect |
|---|---|
| `AccessCreate/Update/Delete/Exists/List` | Basic CRUD |
| `AccessGrantToGroup(group, accesses)` | Grants named rules to group |
| `AccessRevokeFromGroup(group, accesses)` | Revokes named rules from group |
| `GroupAccessList(group)` | Rules currently granted to group |
| `AccountAccessList(accountID)` | All effective rules for account, as `map[string][]*Access`. Each service only receives accesses under its own namespace; the map structure mirrors the IAM interface for UI reuse. |

### TimeControl

`nil` = always valid.

| Field | Description |
|---|---|
| `TimeRanges []TimeRange{Start,End}` | Clock-time windows per repeat period |
| `RepeatFrequency` | `OnlyOnce`(0), `Daily`(1), `Weekly`(2), `Monthly`(3) |
| `RepeatInterval` | Every N periods |
| `RepeatIndexes` | Weekly: day-of-week (1–7); Monthly: day-of-month (1–31) |
| `RepeatEndTime` | Repetition end; zero = no end |
| `IgnoreLoc` | If true, ignore timezone in comparisons |

### Phone Country Code Filtering

`SupportedCountryCodesGet()` → mode + list. Mode `"all"`: no restriction. `"include"`: allow-list. `"exclude"`: deny-list. Validate before login or OTP-send calls.

### System Config

| Method | Effect |
|---|---|
| `RenameEnabled()` | Whether username changes are allowed |
| `ServiceMfaSet(bool)` / `MfaEnforced()` | Service-wide MFA control |
| `JWKSGet()` | JWKS for external token verification |

---

## 11. Subscription

Obtained via `ASNController.GetSubscription()`. Supports Apple App Store, Google Play, Stripe.

### Platform Registration

Register during `Start()`; each returns an HTTP handler for webhook mounting and an `errChan` for backend failures.

```go
appleHandlerFn, errChan, err := sub.AddApple(envConfig, apiConfig)
googleHandlerFn, errChan, err := sub.AddGoogle(envConfig, replayConfig)
stripeHandlerFn, errChan, err := sub.AddStripe(config)
```

Mount handlers via `WebHandler()`. Monitor `errChan` in a background goroutine.

### Notifications

`GetNotificationChannel()` — unified channel emitting a string on any lifecycle event (new, renew, cancel) from any platform.

### Queries

| Method | Behavior |
|---|---|
| `GetUserSubscription(accountID)` | Active subscription or `(nil, false, nil)` |
| `ListUserSubscriptions()` | All active subscriptions |
| `DeleteUserSubscription(accountID)` | Removes local record; does not cancel on platform |
| `RestoreApplePurchaseToken(accountID, token)` | Re-links App Store purchase |
| `RestoreGooglePurchaseToken(accountID, token)` | Re-links Play Store purchase |

### Stripe-Specific

| Method | Returns |
|---|---|
| `GetStripeProductInfo(priceID)` | `*Product{Name, Description, prices by currency}` |
| `GetStripePaymentLink(accountID, priceID, qty, redirectURL)` | Checkout URL |
| `GetStripeBillingPortalUrl(accountID, returnURL)` | Customer Portal URL |

---

## 12. Common Abstractions

### 12.1 Document Database (DocDB)

`DocDBHandler` (from `InitDocDB`): MongoDB and file-based backends.

```
DocDBHandler
  ├── FindOrCreateCollection(name, options) → DocCollection
  ├── DeleteCollection / ListCollections
  ├── CreateFile(fileName, metadata, file)
  └── FindFiles(queryString)
```

Index creation: `DocCollOptions.Indexes`; compound key format: `"col1;col2;..."`.

`DocCollection` operations:

| Method | Cardinality contract |
|---|---|
| `AddRecord` / `AddRecords` | — |
| `FindRecord` / `FindRecordWithRegex` | Errors if 0 or >1 match |
| `FindRecords` / `FindRecordsWithRegex` | Any |
| `UpdateRecord` | Errors if 0 or >1 match |
| `UpdateRecords` | Any |
| `DeleteRecord` | Errors if 0 or >1 match |
| `DeleteRecords` | Any |
| `ArrayAppend(queryJson, newRecord, ignoreDuplicate)` | Appends to array field |
| `ArrayDelete(queryJson, newRecord)` | Removes from array field |
| `ListAllRecord()` | All documents |

Pagination (`page`, `num`): `(-1,≤0)` no limit; `(-1,>0)` limit only; `(≥0,>0)` paginated (0-indexed); `(≥0,≤0)` error.

**Conversion helpers:**

| Function | Direction |
|---|---|
| `MapToStruct[T]` / `MapsToStructs[T]` | `map[string]any` → struct (JSON) |
| `StructToMap[T]` / `StructsToMaps[T]` | struct → `map[string]any` (JSON) |

### 12.2 Time-Series Database (TSDB)

`TSDBHandler` (from `InitTSDB`): InfluxDB v1/v2 and MongoDB TS.

```
TSDBHandler
  ├── FindOrCreateCollection(name, options) → TSCollection
  ├── DeleteCollection / ListCollections
  ├── AddRetentionPolicies / DeleteRetentionPolicies
  └── NewPoint / AddDataPoint(pt, useTimeout) / AddDataPoints(pts, useTimeout)
```

`TSCollection` operations:

| Method | Notes |
|---|---|
| `NewPoint()` / `NewQuery()` | Builder pattern |
| `AddDataPoint` / `AddDataPoints` | Write |
| `FindDataPoints(query)` | Typed query builder → `TSQueryResult` |
| `FindDataPointsBySyntax(syntax)` | Raw DB syntax |
| `FindDataPointsWithRegex(query)` | Regex-filtered |
| `FindPivotDataPoints(query)` | Row-pivoted → `TSPivotQueryResult` |
| `CountDataPoints(query)` | Count without retrieval |

`TSQueryResult`: `Results map[string]TSMatrix`; `TSMatrix` maps tag-group strings to `TSSamples` (ordered `[]TSSamplePoint{Timestamp, Value}`).

### 12.3 Distributed Lock

From `InitLocker()`. **Controller side only.**

```go
Lock(key, identifier string) error
Unlock(key, identifier string) error
```

`identifier` prevents accidental unlock by a different holder.

### 12.4 Logger

From `InitLogger()` on both sides.

| Field | Alias | Purpose |
|---|---|---|
| `Logger.R` | `rlog` / `rtlog` | Runtime — operational events |
| `Logger.A` | `alog` / `apilog` | API access — gRPC, REST, CLI |
| `Logger.P` | `plog` / `perflog` | Performance — latency, throughput |
| `Logger.E` | `elog` / `entitylog` | Entity audit — create/read/update/delete |

Each: `Trace`, `Debug`, `Info`, `Warn`, `Error`, `Fatal`, `Panic` (plain and `f`-format).

### 12.5 Version

Format: `vMAJOR.MINOR[.BUILD[-SUFFIX]]`

```go
v, _ := commonapi.InitVersion("v26.5.4")
v.ToString()      // "v26.5.4"
v.Compare(other)  // VersionCompareGreater(1) | VersionCompareLess(2) | VersionCompareEqual(3)
```

---

## 13. Network and Node Topology

### Network Hierarchy

`GetNetworks()` returns top-level networks; each `Network` embeds child `Networks` via `ParentID`.

`Network.Tiers` — subset of the 12-level location hierarchy:
```
world → country → state → city → district → campus → building → floor → room → row → rack → unit
```

### Node Types (`commonapi.NodeType`)

`router` · `switch` · `appliance` · `firewall` · `lb` · `ap` · `device` · `server`

### Node Hardware Info (`commonapi.NodeInfo`)

| Field | Content |
|---|---|
| `Interfaces map[string]*Interface` | Per-interface IP and `NetIfType` tags |
| `Ipmi *Ipmi` | OOB management credentials |
| `Management *Management` | Hostname and management IP |
| `DeviceInfo *DeviceInfo` | Serial number, vendor, model |
| `DeviceParam *DeviceParam` | CPU cores, memory (bytes), disk (bytes) |

`NetIfType`: `data` · `control` · `management` · `inbound` · `outbound`

### Links (`capi.Link`)

```go
type Link struct {
    ID          string    // UUID
    Description string
    Bandwidth   int64     // bps; symmetric
    From, To    *LinkNode // {NodeID, Interface}
}
```

Internal: both endpoints in queried network; `To` node in result. External: `To` outside network; not in result.

---

## 14. Implementation Checklist

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

- [ ] `SubscribeNodeStateChanges()`: called at most once
- [ ] `runtimeErrChan`: carries only fatal errors, not service-internal errors
- [ ] `opCmd` / `opParams` schemas: defined and shared across controller and service node
- [ ] Shared data keys and value types: agreed upon out-of-band between service teams
- [ ] Config op payload format: versioned for rolling upgrade compatibility
