# ASN Service API

Interface package shared by all services built on the ASN (Amiasys Service Network) distributed framework.  
For the full API contract, see [DESIGN.md](DESIGN.md).

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

## Versioning

A service plugin is a Go plugin: it loads only into an ASN runtime built from
exactly the same version of every package they share — this module, the Go
toolchain, and every other shared module. Any change here, even an additive
one, therefore breaks the plugin ABI. The version policy makes the version
number carry that ABI (design: `ASN.25/design/service_package_dependency.md`).

- **Only `X.Y.0` is released.** There are no patch releases. Every API change
  starts a new minor line.
- **A new line is also required without any Go API change** when the Go
  toolchain version changes, or when the version of any module shared between
  the ASN runtimes and plugins changes (as recorded by the reference `go.mod` on
  the `asn-service-utils` `release/X.Y.0` branch).
- **The minor number is a counter**; it may grow beyond 12 within a year. When
  the major number moves to the next year, the next line is `(X+1).0.0`.
- **The API is final when `X.Y.0` is tagged.** All API work for a line lands
  before the tag; ASN runtime builds on line `X.Y` (runtime version
  `X.Y.<build>`) start only after it, and are built only against `X.Y.0`.

Check a tag before creating it:

```
make check-tag TAG=v26.11.0
```

It rejects anything but `vX.Y.0` and a tag that already exists.

## Getting started

Refer to `asn-service-template` for a minimal working example of both plugins, including build configuration and deployment layout.
