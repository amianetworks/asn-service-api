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
| `log` | `/log` | Logger interface (runtime, API, entity and performance text logs) and structured log records |

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
- The document handler follows the am.modules v27 `docdb` API: updates are
  `DocUpdate` values (`DocSet`, `DocUnset`, `DocInc`, `DocPush`,
  `DocAddToSet`, `DocPull`, `DocPullWhere`, combined with `DocUpdates`), and
  collections add `ReplaceRecord`, `UpsertRecord` and `ReplaceOrAddRecord`.
  The v26 JSON-string `UpdateRecord`/`UpdateRecords`, `ArrayAppend`,
  `ArrayDelete`, `ListAllRecord` (use `FindRecords(DocAll())`) and
  `DocCollOptions.UseRawFormat` are gone.

v27.0.0 is not used: its `NotifyOnClose` value did not match the framework.
Target v27.3.0.

## Moving from v27.3 to v27.4

- Audit: saved filters are removed: `ListAuditFilters`, `SaveAuditFilter`,
  `DeleteAuditFilter`, `SavedAuditFilter`, the `AuditFilterScope*` and
  `AuditPreset*` constants. A service that keeps query templates stores them
  itself.
- Audit: `AuditFilter.SubjectID` becomes `AuditFilter.SubjectIDs`, matching
  any of its subjects. Replace `SubjectID: id` with
  `SubjectIDs: []string{id}`:

  ```go
  records, total, err := ctl.ListAuditRecords(controller.AuditFilter{
      Service:    "myservice",
      SubjectIDs: []string{accountA, accountB},
      Start:      start,
      Stop:       stop,
  }, iam.Page{Limit: 100})
  ```

- Events: a `NotifyRule` selects either by level (`Level`, with `LevelOp`)
  or by event rules (`EventRuleIDs`, at least one, with `Level` and `LevelOp`
  empty), never both; both or neither is `shared.request.param_invalid`. An
  empty `EventRuleIDs` no longer means "every rule" next to a level: a rule
  by level reads every rule of the namespace and root. To keep only some
  levels of one event rule, put a filter on its level tag in the event rule.

  ```go
  byRules := controller.NotifyRule{Name: "disk", RootID: root,
      EventRuleIDs: []string{diskRuleID}, SendOn: []string{controller.NotifyOnOpen},
      ContactGroupIDs: []string{groupID}}
  byLevel := controller.NotifyRule{Name: "high", RootID: root,
      LevelOp: controller.NotifyLevelGE, Level: "high", SendOn: []string{controller.NotifyOnOpen},
      ContactGroupIDs: []string{groupID}}
  ```

- Catalog: `ServiceCatalog.Service` is removed. ASN knows a service by the
  name it loaded the plugin with (the `.so` file name, which is the
  service's name in `asn.conf`). A data item is declared by its source
  alone, not empty and with no `.`; ASN stores and shows it as
  `<service>.<source>`. A default rule leaves `Service` empty and names
  items by source.

  ```go
  // v27.3
  catalog := &controller.ServiceCatalog{Service: "myservice", /* ... */
      Items:        []controller.CatalogItem{{Name: "myservice.probe", /* ... */}},
      DefaultRules: []controller.EventRule{{Name: "slow", Item: "myservice.probe", /* ... */}}}
  // v27.4
  catalog := &controller.ServiceCatalog{ /* ... */
      Items:        []controller.CatalogItem{{Name: "probe", /* ... */}},
      DefaultRules: []controller.EventRule{{Name: "slow", Item: "probe", /* ... */}}}
  ```

- Events: on the service API a service names its own items by source and
  every other item in full (`asn.host`, another service's
  `<service>.<source>`), in points, rules, conditions and the `Items`
  filters; ASN returns its own items by source the same way (rules,
  incidents, `DataDemand.Item`, its own catalog from `GetServiceCatalog`).
  Naming an own item in full (`myservice.probe`) is
  `shared.request.param_invalid`. A rule, incident, notify rule, contact
  group or send of the service's own namespace reads back with `Service`
  empty; `Service` still takes the service's own name as its own namespace.
  Code that tells its own records by name (`inc.Service == "myservice"`)
  checks for empty instead.

  ```go
  err := ctl.WriteEventPoints("", []controller.EventPoint{{
      Item:   "probe", // was "myservice.probe"
      Tags:   map[string]string{"NodeID": nodeID},
      Fields: map[string]any{"latency_ms": 12.5},
  }})
  for inc := range incidents {
      if inc.Service == "" && inc.Item == "probe" { /* own */ }
      if inc.Service == "asn" && inc.Item == "asn.host" { /* framework */ }
  }
  ```

- Time series: `TSCollOptions.Rollups` (`TSRollup`, `TSFieldType`) and
  `TSCollOptions.SeriesLookup` speed up `Aggregate` and reads by tags, with
  the same results. They apply when a collection is created; for a
  collection that exists, call `TSDBHandler.BuildIndexes` (series lookup,
  new indexes) and `TSDBHandler.BuildRollups` (rollups) once, for example
  from `Start()`, in a goroutine: both run without timeout and do nothing
  the second time. Use the series lookup for collections read a few series
  at a time (one node), not for many series of a few points each. A type
  that implements `TSDBHandler` adds the two methods.

  ```go
  opts := &commonapi.TSCollOptions{Retention: 30 * 24 * time.Hour,
      Indexes:      []commonapi.TSIndex{{Tags: []string{"NodeID"}}},
      Rollups:      []commonapi.TSRollup{{Window: time.Hour, Fields: map[string]commonapi.TSFieldType{"rx": commonapi.TSFieldFloat}}},
      SeriesLookup: true}
  coll, err := ts.FindOrCreateCollection("traffic", opts) // new collections get both
  go func() {                                            // existing ones, once
      if err := ts.BuildIndexes("traffic", opts); err != nil { /* log */ }
      if err := ts.BuildRollups("traffic", opts); err != nil { /* log */ }
  }()
  ```

## Moving from v27.2 to v27.3

- Audit: `AuditCatalog.TagKeys` is `[]AuditTagKey`: each `Key` (same rules
  as before) has display `Names` (at least one locale) and optional `Values`
  (`AuditTagValue{Value, Names}`, at most `AuditTagValuesMax`), for display
  only. Change a declaration `TagKeys: []string{"Role"}` to
  `TagKeys: []AuditTagKey{{Key: "Role", Names: map[string]string{"zh-CN": "角色", "en-US": "Role"}}}`.
- Log parsers: optional `LogParser.Fields` (`LogParserField{Name, Label,
  Type}`) names fields of the patterns for display; `Type` is read only.
- Capture: `SnapLen` up to `CaptureSnapLenMax` (2048) with
  `MaxPackets × SnapLen ≤ CaptureWindowBytesMax`; `CaptureMaxPacketsMax`
  is exported. Fields and types of `CaptureSettings` are unchanged.
- Default rules: a catalog's default rule may set `Scope` to
  `EventScopePlatform`: ASN installs it once in the platform scope instead of
  in each root. It reads items that are not node data. An empty `Scope` (or
  `EventScopeRoot`) installs in each root as before.
- Comments: `PacketMeta` and `CaptureWindow.Headers`, `AccountCreate` and
  `AccountDelete`, the syslog receiver's format detection.
- Only `TagKeys` declarations need a source change; every plugin is rebuilt
  on v27.3.0.

## Moving from v27.1 to v27.2

- Audit: `AuditFilter.DeviceID` and `AuditFilter.Tags`; `AuditCatalog.TagKeys`
  declares the keys of the service's own audit tags it queries by (at most
  `AuditTagKeysMax`, none of `AuditReservedTagKeys`, checked by
  `ValidateServiceCatalog`); `AuditSummary.ByTag`. `WriteAudit` accepts a
  root or node deleted in the last 10 minutes.
- Structured logs: `StructuredLogAPI` (embedded in `ASNController`) keeps the
  service's own entity log (`log.EntityLog`) and API log (`log.APILog`),
  apart from other services and from the operation audit. `AuditAPI`
  stays the record operators review across services.
- Source using v27.1 compiles unchanged; every plugin is rebuilt on v27.2.0.

## Getting started

Refer to `asn-service-template` for a minimal working example of both plugins, including build configuration and deployment layout.
