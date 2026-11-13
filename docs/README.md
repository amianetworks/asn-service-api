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

## Moving from v27.5 to v27.6

Five changes affect the compiled API: `EventsAPI.EventRetention`,
`AuditSubjectCount.Username`, `TSCollection.FindPage` and
`TSCollection.CountWindows`, and the `byAdmin` parameter and `existed` result of
`iam.Instance.AccountCreate`. The rest documents rules ASN already applies.

- Time series: `TSCollection` gains `FindPage` and `CountWindows`, the
  am.modules `tsdb.Collection` methods of the same names, so that a service
  adapting the handler to `tsdb.Database` (for the am.modules audit, log and
  event modules) passes them through. `FindPage` returns a page of the points
  having a field, newest first; `CountWindows` counts them per window of the
  range and per tag values, returning every window (`TSWindow`), empty ones
  without groups. A type that implements `TSCollection` adds the two methods.

  ```go
  wins, err := coll.CountWindows(commonapi.TSQuery{Start: from, Stop: to},
      "action", time.Hour, "result")
  for _, w := range wins {
      for _, g := range w.Groups {
          _ = g.Values["result"] // g.Count points in [w.Start, w.Start+1h)
      }
  }
  ```

- Events: `EventRetention()` returns how long incidents are kept, set by the
  deployment (1–3650 days, 365 by default). A notify rule's `Cooldown` and
  `MinInterval` are at most this: `CreateNotifyRule` and `UpdateNotifyRule`
  refuse a longer one with `shared.request.param_invalid`. A form can bound
  its inputs before the call:

  ```go
  limit := ctl.EventRetention()
  if rule.Cooldown > limit || rule.MinInterval > limit {
      // tell the user the most is limit
  }
  ```

- Events: `asn.event.rule_in_use` from `DeleteEventRule` gives the notify
  rules naming the rule in `errMeta.notify_rules`, their IDs sorted and
  comma-separated.
- Audit: `SummarizeAudit` fills `AuditSubjectCount.Username` for an account
  subject: the username its latest record names. It is empty for system and
  service subjects. Show it in place of the account ID:

  ```go
  sum, err := ctl.SummarizeAudit(filter, time.Hour, 10)
  for _, s := range sum.TopAlarmSubjects {
      who := s.SubjectID
      if s.Username != "" {
          who = s.Username
      }
      _ = who
  }
  ```

- Audit: `WriteAudit` of an `account` subject needs its `SubjectID` (the
  account ID); without one it is `shared.request.param_invalid`. A `service`
  subject leaves `SubjectID` empty and ASN sets the caller.
- Audit: `ExportAuditRecords` writes newest first, as `ListAuditRecords`
  lists.
- IAM: `AccountCreate` takes `byAdmin` and returns `existed`. Accounts are
  global, so a username that already exists outside the service names that
  account, and the call adds it to the service with its existing ID.
  - `byAdmin` true, for an administrator creating the account: an existing
    account is added as it is, keeping its own password, email and phone;
    nothing in the request is checked against it. To create a different person,
    use another username.
  - `byAdmin` false, for self-registration: the person proves the account is
    theirs by its password or by a code sent to the account's own email or
    phone. Without a username, the account is found by its email, or else by its
    phone. After a code proof, the password in the request becomes the
    account's password and the account's sessions end, as in a password
    recovery.

  Tell the administrator when an account was added rather than created:

  ```go
  iamInst := ctl.GetIAM()
  id, existed, err := iamInst.AccountCreate(username, "", password,
      email, "", true, phone, "", true, true)
  if err == nil && existed {
      // "username already exists; the account was added and keeps its own
      // password, email and phone"
  }
  ```

  A self-registration form sends a code to the email and passes it:

  ```go
  // after AccountEmailCodeSend(email, false)
  id, existed, err := iamInst.AccountCreate("", "", newPassword,
      email, code, false, nil, "", true, false)
  ```

- IAM: after `Logout`, the device's refresh token is refused at once and its
  access token once IAM syncs the revocation, within
  `authentication.revoke_sync_interval` (1 s by default). `TokenRevoke` takes
  effect the same way. A service does not call `TokenRevoke` after `Logout`.

## Moving from v27.4 to v27.5

- Events: `ContactGroup.NotifyRules` counts the notify rules naming the
  group, in `ListContactGroups` and in what `CreateContactGroup` and
  `UpdateContactGroup` return; it is ignored on write. A group with any
  cannot be deleted (`DeleteContactGroup` is `shared.resource.in_use`), so a
  list can tell before the call:

  ```go
  groups, total, err := ctl.ListContactGroups(ctl.ContactGroupFilter{RootID: root}, iam.Page{Limit: 100})
  for _, g := range groups {
      canDelete := g.NotifyRules == 0
      _ = canDelete
  }
  ```

- Events: `IncidentFilter.State` is replaced by `States` (any of them);
  `IDs` selects up to 1000 incidents, as the management API's `states` and
  `ids`. An unknown state, more than 1000 IDs or a bad ID is
  `shared.request.param_invalid` (`errMeta.field` is `states` or `ids`):

  ```go
  incs, total, err := ctl.ListIncidents(ctl.IncidentFilter{RootID: root,
      States: []string{ctl.IncidentOpen, ctl.IncidentRecorded}}, iam.Page{Limit: 50})
  // The incidents a dispatch mailed, a page at a time:
  merged, _, err := ctl.ListIncidents(ctl.IncidentFilter{RootID: root, IDs: dispatch.IncidentIDs},
      iam.Page{Limit: 500})
  ```

- Events: the lists read several namespaces at once. `IncidentFilter` and
  `EventRuleFilter` take `Services []string` in place of `Service`;
  `ListNotifyRules` and `ListContactGroups` take a filter
  (`NotifyRuleFilter`, `ContactGroupFilter`) with `Services` in place of
  the `service, rootID` parameters. Nil or empty is the caller's own; `""`
  or the caller's name in the list is its own. The namespaces are sorted,
  paged and counted as one list; a namespace the caller cannot read is
  `asn.event.namespace`. A service that is not super reads in `asn` the
  rules bound to it and the incidents of its nodes as before, but no longer
  the incidents of rules on `asn.audit` (`GetIncident` and
  `ResolveIncident` of one are `shared.resource.not_found`,
  `SubscribeIncidents` does not send them), and no `asn` notify rules or
  contact groups. `ListIncidents` counts at most `IncidentTotalMax` (10000):
  a total of it is that many or more. `ListDeliveries` keeps one `Service`:

  ```go
  // Its own incidents and the asn incidents of its nodes, in one list:
  incs, total, err := ctl.ListIncidents(ctl.IncidentFilter{Services: []string{"", "asn"},
      RootID: root, States: []string{ctl.IncidentOpen}}, iam.Page{Limit: 50})
  ```

- Events: list conditions. `EventRuleFilter.Levels` selects rules of a
  fixed level (a rule whose level comes from its data matches none) and
  `Builtin` the rules installed from a catalog or the others.
  `NotifyRuleFilter` has `Keyword` (the name), `Enable` and
  `ContactGroupID` (the rules mailing a group); `ContactGroupFilter.Keyword`
  matches the name or an address of `To` or `Cc`. Keywords are contained,
  case-insensitive:

  ```go
  builtin := true
  rules, _, err := ctl.ListEventRules(ctl.EventRuleFilter{RootID: root,
      Levels: []string{ctl.EventLevelHigh, ctl.EventLevelCritical}, Builtin: &builtin}, iam.Page{Limit: 50})
  nrs, _, err := ctl.ListNotifyRules(ctl.NotifyRuleFilter{RootID: root, ContactGroupID: groupID},
      iam.Page{Limit: 50})
  ```

- Events: references carry names. `Ref{ID, Name}` replaces the IDs:
  `EventRule.LabelID` and `NodeIDs` are `Label *Ref` and `Nodes []Ref`;
  `NotifyRule.EventRuleIDs`, `ContactGroupIDs` and `NodeIDs` are
  `EventRules`, `ContactGroups` and `Nodes []Ref`. A write reads the IDs;
  a read sets each `Name`, empty when the object no longer exists.
  `Incident.NodeName`, `AuditRecord.NodeName` and `LogRecord.NodeName` are
  the node's name; `AuditRecord.Username` is the account's name when the
  record was written. `GetNodesByIDs` reads up to `GetNodesByIDsMax` (1000)
  nodes, with their names and labels, in one call:

  ```go
  nr := ctl.NotifyRule{Name: "disk", RootID: root, EventRules: []ctl.Ref{{ID: diskRuleID}},
      SendOn: []string{ctl.NotifyOnOpen}, ContactGroups: []ctl.Ref{{ID: groupID}}}
  got, err := ctl.CreateNotifyRule(nr) // got.EventRules[0].Name is the rule's name
  nodes, err := ctl.GetNodesByIDs(nodeIDs) // the nodes the caller reads, in order
  ```

- Events: `ExportIncidents(filter, format, w)` writes what `ListIncidents`
  reads as `csv` or `jsonl`, newest first, at most 10000, with `truncated`
  when more matched; the columns are the management API export's. It is
  audited. Who among the service's users may export is the service's to
  decide:

  ```go
  rows, truncated, err := ctl.ExportIncidents(ctl.IncidentFilter{RootID: root}, "csv", w)
  ```

- Objects shown with actions say what the calling service may do with
  them now, read only, set by ASN with the checks the write makes:
  `Writable` and `Deletable` on `EventRule`, `NotifyRule`, `ContactGroup`,
  `Label` and `LogParser`, `Resolvable` on `Incident`. Counts beside them
  say why an object is not deletable, named as the refusal's `errMeta`:
  `EventRule.NotifyRules`, `ContactGroup.NotifyRules`, `Label.Rules` and
  `Label.Bindings`; `Builtin` stays. A page grays an action out from them
  instead of working the rules out itself; the write still decides, and
  which of the service's users may act stays the service's call:

  ```go
  for _, r := range rules {
      switch {
      case !r.Writable:
          // no access: edit and delete grayed
      case r.Builtin:
          // built in: delete grayed
      case r.NotifyRules > 0:
          // named by notify rules: delete grayed
      }
      _ = r.Deletable // true exactly when none of the above
  }
  ```

- Events: `DeleteEventRule` (`asn.event.rule_in_use`) and
  `DeleteContactGroup` (`shared.resource.in_use`) refusals carry
  `errMeta.notify_rules`: the IDs of the notify rules naming the object,
  comma-separated.
- Catalog: `ValidateServiceCatalog` refuses a default rule of one instant
  condition handled `auto` or `any` (`DefaultRules[i].handling`), as ASN
  refuses it when it installs the rule. Such a rule is handled `none`,
  `manual` or left empty (none):

  ```go
  {Name: "port-down", Item: "port", Level: "high", Handling: ctl.EventHandleNone,
      Conditions: []ctl.EventCondition{{Mode: ctl.EventModeInstant, Tag: "State", Op: "eq", Values: []string{"down"}}}}
  ```

- Events: a rule has no scope. `EventRule.Scope`, `EventRuleFilter.Scope`
  and the `EventScope*` constants are removed, and so are rules of every
  root: a rule belongs to its `RootID`, a root or `commonapi.PlatformScope`
  for the points of no root. A rule wanted in several roots is created in
  each. `EventRuleFilter.Devices` is removed too: `NodeID` finds the rules
  that cover a device.
- Catalog: a default rule no longer sets `Scope`; the item decides. An item
  whose points are written with no root declares `CatalogItem.Platform`
  (not with `Node`), and a default rule on it is installed once for the
  platform; any other default rule is installed in each root of the
  service's nodes. The items one default rule reads are all `Platform` or
  none:

  ```go
  Items: []ctl.CatalogItem{{Name: "license", Demand: ctl.CatalogDemandAlways, Platform: true,
      Fields: []ctl.CatalogItemField{{Name: "usage", Type: ctl.CatalogItemNumber}}}},
  DefaultRules: []ctl.EventRule{{Name: "license-high", Item: "license", Level: ctl.EventLevelHigh,
      Conditions: []ctl.EventCondition{{Mode: ctl.EventModeThreshold, Field: "usage", Op: "gt", Value: 90}}}},
  ```

- Audit: `AuditFilter.EventTypes`, `Operations` and `ObjectTypes` need
  `Service`, as `Tags` do: they are names and codes of one service's audit
  catalog. Without it the query is `shared.request.param_invalid`
  (`errMeta.field` is `event_types`, `operations` or `object_types`).
- Audit: `AuditFilter.Usernames` selects records by the account name
  recorded with them (exact match, any of them), for a page that knows the
  name and not the account ID. `AuditFilter.Origins` is removed: the
  record's `Origin` is still returned:

  ```go
  recs, total, err := ctl.ListAuditRecords(ctl.AuditFilter{Usernames: []string{"alice"},
      Start: time.Now().Add(-24 * time.Hour), Stop: time.Now()}, iam.Page{Limit: 50})
  ```

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
