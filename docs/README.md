# ASN Service API

Interface package shared by all services built on the ASN (Amiasys Service Network) distributed framework.  
For the full API contract, see [ASN_SERVICE_API_DESIGN.md](ASN_SERVICE_API_DESIGN.md).

## Module

```
asn.amiasys.com/asn-service-api/v26
```

## Package Layout

| Package | Import suffix | Role |
|---|---|---|
| `capi` | `/controller` | Controller-side interfaces and structs |
| `snapi` | `/servicenode` | Service-node-side interfaces and structs |
| `commonapi` | `/common` | Shared enums, structs, DB/log abstractions |
| `iam` | `/iam` | IAM interface |
| `subscription` | `/subscription` | In-App Purchase / Subscription interface |
| `log` | `/log` | Structured logger interface |

## What to Implement

A service consists of two independently loaded `.so` plugins: one for the controller and one for each service node.

### Controller plugin

Implement `capi.ASNServiceController` (defined in `controller/service.go`) and export the constructor:

```go
func NewASNServiceController() capi.ASNServiceController
```

The framework calls this function to instantiate the controller plugin. The function name, parameter list, and return type must match exactly.

### Service node plugin

Implement `snapi.ASNService` (defined in `servicenode/service.go`) and export the constructor:

```go
func NewASNService() snapi.ASNService
```

The framework calls this function to instantiate the service node plugin. The function name, parameter list, and return type must match exactly.

## Framework-provided handles

The framework passes its own implementations to your code during `Init()`:

- `capi.ASNController` — passed to `ASNServiceController.Init()`; provides topology queries, service lifecycle management, ops dispatch, IAM, and subscriptions.
- `snapi.ASNServiceNode` — passed to `ASNService.Init()`; provides node information, DB/log handles, and cross-service data access.

Do not implement these interfaces; only consume them.

## Getting started

Refer to `asn-service-template` for a minimal working example of both plugins, including build configuration and deployment layout.
