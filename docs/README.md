# ASN Service API

Interface package shared by all services built on the ASN (Amiasys Service Network) distributed framework.  
For the full API contract, see [DESIGN.md](DESIGN.md).

## Module

```
asn.amiasys.com/asn-service-api/v27
```

## Package Layout

| Package | Import suffix | Role |
|---|---|---|
| `capi` | `/controller` | Controller-side interfaces and structs |
| `snapi` | `/servicenode` | Service-node-side interfaces and structs |
| `commonapi` | `/common` | Shared enums, structs, DB/log abstractions |
| `iam` | `/iam` | IAM interface |
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

- `capi.ASNController` — passed to `ASNServiceController.Init()`; provides topology queries, service lifecycle management, ops dispatch, and IAM.
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

## Moving from v26 to v27

- Import path `asn.amiasys.com/asn-service-api/v27`.
- `StaticResource.CLICommands(cli CLIContext)`: `cli.ApplyOps` is the former
  function argument.
- `StaticResource.Catalog()`: nil declares no catalog.
- `GetMetricsTemplate(target)` and `GetMetrics(target, window)`: an empty
  template for a node target and a template without windows suffice.
- `OnDataDemand(demands)`: return nil when no item is declared on demand.
- `serviceName` (`""` for the caller) is the first argument of the
  cross-service methods.
- The time-series handler: `Find`, `Count`, `CountBy`, `Scan`, `FindAfter`
  and `Aggregate` on `TSQuery` / `TSAggregation`; retention is set per
  collection in `TSCollOptions`. Existing data is migrated by the service.
- New capabilities on `ASNController`: labels, audit, events, host metrics,
  logs, node attributes and risk, catalogs and metrics of other providers
  (see DESIGN.md §5.2).

## Getting started

Refer to `asn-service-template` for a minimal working example of both plugins, including build configuration and deployment layout.
