// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// The service catalog: what a service declares so that ASN, ASN Web and
// super services understand it without code written for it. A controller
// plugin returns it from StaticResource.Catalog(); ASN validates it with
// ValidateServiceCatalog at load and refuses the plugin when it is invalid.
// Every service reads any catalog with ASNController.GetServiceCatalog,
// "asn" included.
//
// A catalog does not name its service: ASN knows the service by the name
// it loaded the plugin with. A data item is declared by its source alone;
// ASN stores and shows it as <service>.<source>. On the service API a
// service names its own items by source and every other item (another
// service's, "asn"'s) in full, and reads its own items back by source.

// Grouping values: whether the service uses ASN node groups.
const (
	CatalogGroupingASN  = "asn-groups"
	CatalogGroupingNone = "none"
)

// Demand values of a data item: written always, or only while a rule reads
// it (see ASNServiceController.OnDataDemand).
const (
	CatalogDemandAlways   = "always"
	CatalogDemandOnDemand = "onDemand"
)

// Data item field and tag types.
const (
	CatalogItemNumber = "number"
	CatalogItemEnum   = "enum"
	CatalogItemText   = "text"
)

// Configuration schema field types.
const (
	CatalogFieldString   = "string"
	CatalogFieldInt      = "int"
	CatalogFieldBool     = "bool"
	CatalogFieldEnum     = "enum"
	CatalogFieldDuration = "duration"
	CatalogFieldIP       = "ip"
	CatalogFieldCIDR     = "cidr"
	CatalogFieldList     = "list"
	CatalogFieldObject   = "object"
	CatalogFieldRaw      = "raw"
)

var catalogFieldTypes = []string{CatalogFieldString, CatalogFieldInt, CatalogFieldBool, CatalogFieldEnum,
	CatalogFieldDuration, CatalogFieldIP, CatalogFieldCIDR, CatalogFieldList, CatalogFieldObject, CatalogFieldRaw}

// CatalogDisplay is how a service is shown.
type CatalogDisplay struct {
	Name  map[string]string // language (zh-CN, en-US) -> name; at least one
	Icon  string
	Order int
	URL   string // the service's own UI, if any
}

// CatalogField is a configuration schema field.
type CatalogField struct {
	Name            string
	Type            string
	Required        bool
	Min, Max        *float64 // numbers; lengths of strings and lists
	Values          []string // enum
	Elem            *CatalogField
	Fields          []CatalogField // object
	RestartRequired bool
	Label           map[string]string
}

// CatalogSchema is a configuration schema; Raw takes any YAML value.
type CatalogSchema struct {
	Raw    bool
	Fields []CatalogField
}

// CatalogItemField is a field or tag of a data item.
type CatalogItemField struct {
	Name   string
	Type   string // number, enum, text; a tag is enum or text, empty for text
	Unit   string
	Values []string
	Label  map[string]string
}

// CatalogItem is a data item: the points of measurement a service writes
// with WriteEventPoints. Name is the source, not empty and with no '.';
// ASN stores and shows the item as <service>.<Name>.
type CatalogItem struct {
	Name   string
	Demand string
	Node   bool // its points are about a node and carry NodeID
	Fields []CatalogItemField
	Tags   []CatalogItemField
	Label  map[string]string
}

// ServiceCatalog is one service's catalog. It does not name the service.
type ServiceCatalog struct {
	Display       CatalogDisplay
	Grouping      string
	Config        *CatalogSchema // nil: the service declares no configuration
	Items         []CatalogItem
	PacketCapture bool
	Audit         *AuditCatalog
	// DefaultRules are event rules ASN installs, disabled and built in, in
	// the service's namespace, once: a rule with Scope empty (or
	// EventScopeRoot) in each root network of the service's nodes; a rule
	// with Scope EventScopePlatform once in the platform scope, while the
	// service is loaded, whether or not it has nodes. A platform rule reads
	// items that are not node data (CatalogItem.Node false): their points
	// written with no root are of the platform scope. A rule installed is
	// never installed again by name, so changing a released rule's Scope
	// leaves the copies installed before. A default rule's Service is
	// empty; its Item and Conditions[].Item are names of Items.
	DefaultRules []EventRule
}

// Item returns the item of a name, nil when it is not declared.
func (c *ServiceCatalog) Item(name string) *CatalogItem {
	if c == nil {
		return nil
	}
	for i := range c.Items {
		if c.Items[i].Name == name {
			return &c.Items[i]
		}
	}
	return nil
}

// Field returns a declared field of the item.
func (it *CatalogItem) Field(name string) *CatalogItemField { return findItemField(it.Fields, name) }

// Tag returns a declared tag of the item.
func (it *CatalogItem) Tag(name string) *CatalogItemField { return findItemField(it.Tags, name) }

func findItemField(fs []CatalogItemField, name string) *CatalogItemField {
	for i := range fs {
		if fs[i].Name == name {
			return &fs[i]
		}
	}
	return nil
}

// CatalogError is a catalog validation failure at Path.
type CatalogError struct {
	Path   string
	Reason string
}

func (e *CatalogError) Error() string { return "catalog " + e.Path + ": " + e.Reason }

func catalogInvalid(path, reason string) error { return &CatalogError{Path: path, Reason: reason} }

// CatalogReservedTags are set by ASN on every point.
var CatalogReservedTags = []string{"Service", "RootID", "NodeID"}

// CatalogReservedPrefixes start ASN's presence tags.
var CatalogReservedPrefixes = []string{"L_", "S_"}

// PointName checks a tag or field name against the store's point rules
// (am.modules 415d788).
func CatalogPointName(name string) string {
	switch {
	case name == "":
		return "empty"
	case strings.Contains(name, "."):
		return "has a '.'"
	case strings.HasPrefix(name, "_") || strings.HasPrefix(name, "$"):
		return "starts with '_' or '$'"
	case name == "time" || name == "result" || name == "table":
		return "is reserved by the store"
	}
	return ""
}

// ValidateServiceCatalog checks a catalog: names, types, item and field rules.
func ValidateServiceCatalog(c *ServiceCatalog) error {
	if c == nil {
		return catalogInvalid("", "no catalog")
	}
	if c.Grouping != CatalogGroupingASN && c.Grouping != CatalogGroupingNone {
		return catalogInvalid("grouping", "grouping is asn-groups or none")
	}
	if len(c.Display.Name) == 0 {
		return catalogInvalid("display.name", "a display name in at least one language")
	}
	if c.Config != nil && !c.Config.Raw {
		if err := validateCatalogFields("config", c.Config.Fields); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for i, it := range c.Items {
		path := fmt.Sprintf("items[%d]", i)
		if it.Name == "" || strings.Contains(it.Name, ".") {
			return catalogInvalid(path+".name", "an item is named by its source: not empty, no '.'")
		}
		if seen[it.Name] {
			return catalogInvalid(path+".name", "item "+it.Name+" declared twice")
		}
		seen[it.Name] = true
		if it.Demand != CatalogDemandAlways && it.Demand != CatalogDemandOnDemand {
			return catalogInvalid(path+".demand", "demand is always or onDemand")
		}
		if len(it.Fields) == 0 {
			return catalogInvalid(path+".fields", "an item has at least one field")
		}
		names := map[string]string{}
		for kind, list := range map[string][]CatalogItemField{"fields": it.Fields, "tags": it.Tags} {
			for j, f := range list {
				fp := fmt.Sprintf("%s.%s[%d]", path, kind, j)
				if why := CatalogPointName(f.Name); why != "" {
					return catalogInvalid(fp, f.Name+" "+why)
				}
				if slices.Contains(CatalogReservedTags, f.Name) || slices.ContainsFunc(CatalogReservedPrefixes, func(p string) bool { return strings.HasPrefix(f.Name, p) }) {
					return catalogInvalid(fp, f.Name+" is set by ASN")
				}
				if other, ok := names[f.Name]; ok {
					return catalogInvalid(fp, f.Name+" is declared in "+other+" too")
				}
				names[f.Name] = kind
				switch f.Type {
				case CatalogItemNumber, CatalogItemText:
				case CatalogItemEnum:
					if len(f.Values) == 0 {
						return catalogInvalid(fp, "an enum has values")
					}
				default:
					if kind == "tags" && f.Type == "" {
						continue
					}
					return catalogInvalid(fp, "type is number, enum or text")
				}
				if kind == "tags" && f.Type == CatalogItemNumber {
					return catalogInvalid(fp, "a tag is enum or text")
				}
			}
		}
	}
	if err := validateAuditCatalog(c.Audit); err != nil {
		return err
	}
	ruleNames := map[string]bool{}
	for i, r := range c.DefaultRules {
		path := fmt.Sprintf("defaultRules[%d]", i)
		switch {
		case r.Name == "" || ruleNames[r.Name]:
			return catalogInvalid(path+".name", "a default rule has a unique name")
		case r.ID != "" || r.RootID != "" || len(r.NodeIDs) > 0 || r.LabelID != "" || r.Bound != "":
			return catalogInvalid(path, "a default rule names no ID, root, nodes, label or bound service")
		case r.Service != "":
			return catalogInvalid(path+".service", "a default rule names no namespace: it is the service's")
		case len(r.Conditions) == 0:
			return catalogInvalid(path+".conditions", "a default rule has conditions")
		}
		if err := validateDefaultRule(c, path, &r); err != nil {
			return err
		}
		ruleNames[r.Name] = true
	}
	return nil
}

var (
	eventLevels   = []string{EventLevelInfo, EventLevelWarn, EventLevelLow, EventLevelMedium, EventLevelHigh, EventLevelCritical}
	eventHandling = []string{"", EventHandleNone, EventHandleAuto, EventHandleManual, EventHandleAny}
	eventOps      = []string{"gt", "ge", "lt", "le", "eq", "ne"}
)

// validateDefaultRule checks a default rule as ASN checks a rule when it is
// created: the item and the fields and tags it reads are declared by the
// catalog, and the level, handling and conditions are well formed.
func validateDefaultRule(c *ServiceCatalog, path string, r *EventRule) error {
	item := c.Item(r.Item)
	if item == nil {
		return catalogInvalid(path+".item", "item "+r.Item+" is not declared")
	}
	switch {
	case r.Scope != "" && r.Scope != EventScopeRoot && r.Scope != EventScopePlatform:
		return catalogInvalid(path+".scope", "a default rule's scope is empty, root or platform")
	case r.Devices != "" && r.Devices != EventDevicesAll:
		return catalogInvalid(path, "a default rule covers every device")
	case r.Scope == EventScopePlatform && item.Node:
		return catalogInvalid(path+".item", "a platform default rule reads no node data: "+r.Item+" is")
	case !slices.Contains(eventLevels, r.Level) && r.LevelFrom == nil:
		return catalogInvalid(path+".level", "level "+r.Level+" is not a level")
	case r.Level != "" && !slices.Contains(eventLevels, r.Level):
		return catalogInvalid(path+".level", "level "+r.Level+" is not a level")
	case !slices.Contains(eventHandling, r.Handling):
		return catalogInvalid(path+".handling", "handling "+r.Handling+" is not none, auto, manual or any")
	}
	if lf := r.LevelFrom; lf != nil {
		if (lf.Tag == "") == (lf.Field == "") {
			return catalogInvalid(path+".levelFrom", "a level comes from a tag or a field")
		}
		for _, l := range lf.Map {
			if !slices.Contains(eventLevels, l) {
				return catalogInvalid(path+".levelFrom.map", l+" is not a level")
			}
		}
		if lf.Default != "" && !slices.Contains(eventLevels, lf.Default) {
			return catalogInvalid(path+".levelFrom.default", lf.Default+" is not a level")
		}
	}
	for _, f := range r.Filters {
		if item.Tag(f.Tag) == nil {
			return catalogInvalid(path+".filters", "tag "+f.Tag+" is not declared")
		}
	}
	for i, cond := range r.Conditions {
		at := fmt.Sprintf("%s.conditions[%d]", path, i)
		ci := item
		if cond.Item != "" {
			if ci = c.Item(cond.Item); ci == nil {
				return catalogInvalid(at+".item", "item "+cond.Item+" is not declared")
			}
			if r.Scope == EventScopePlatform && ci.Node {
				return catalogInvalid(at+".item", "a platform default rule reads no node data: "+cond.Item+" is")
			}
		}
		if cond.Mode == EventModeInstant && len(r.Conditions) > 1 {
			return catalogInvalid(at+".mode", "an instant condition is a rule's only condition")
		}
		if cond.Tag != "" && ci.Tag(cond.Tag) == nil {
			return catalogInvalid(at+".tag", "tag "+cond.Tag+" is not declared")
		}
		if cond.Field != "" && ci.Field(cond.Field) == nil {
			return catalogInvalid(at+".field", "field "+cond.Field+" is not declared")
		}
		if err := validateCondition(at, &cond); err != nil {
			return err
		}
	}
	return nil
}

// validateCondition checks a condition's mode, operator, window and the
// parameters of its mode.
func validateCondition(at string, c *EventCondition) error {
	if c.MaxGap < 0 || (c.MaxGap > 0 && c.Mode != EventModeThreshold && c.Mode != EventModeDuration) {
		return catalogInvalid(at+".maxGap", "only threshold and duration conditions take a max gap, not negative")
	}
	if c.Tag != "" || len(c.Values) > 0 {
		switch c.Mode {
		case EventModeThreshold, EventModeInstant:
			if c.Window != 0 {
				return catalogInvalid(at+".window", "a "+c.Mode+" condition takes no window")
			}
		case EventModeDuration, EventModeRatio, EventModeCount:
			if c.Window <= 0 {
				return catalogInvalid(at+".window", "a "+c.Mode+" condition needs a positive window")
			}
			if c.Mode == EventModeRatio && (c.Percent <= 0 || c.Percent > 100) {
				return catalogInvalid(at+".percent", "percent is in (0, 100]")
			}
			if c.Mode == EventModeCount && c.Count <= 0 {
				return catalogInvalid(at+".count", "count is positive")
			}
		default:
			return catalogInvalid(at+".mode", "a "+c.Mode+" condition compares numbers, not values")
		}
		switch {
		case (c.Tag == "") == (c.Field == ""):
			return catalogInvalid(at+".tag", "a condition on values compares a tag or a field")
		case len(c.Values) == 0:
			return catalogInvalid(at+".values", "a condition on a tag has values")
		case c.Op != "eq" && c.Op != "ne":
			return catalogInvalid(at+".op", "a condition on values takes eq or ne")
		}
		return nil
	}
	needField, needOp, needWindow := true, true, true
	switch c.Mode {
	case EventModeThreshold:
		needWindow = false
	case EventModeInstant:
		needField, needOp, needWindow = false, c.Field != "", false
	case EventModeDuration:
	case EventModeAggregate:
		if !slices.Contains([]string{"avg", "min", "max", "sum", "last"}, c.Aggregate) {
			return catalogInvalid(at+".aggregate", "aggregate is avg, min, max, sum or last")
		}
	case EventModeCount:
		if c.Count <= 0 {
			return catalogInvalid(at+".count", "count is positive")
		}
		needField, needOp = false, c.Field != ""
	case EventModeRatio:
		if c.Percent <= 0 || c.Percent > 100 {
			return catalogInvalid(at+".percent", "percent is in (0, 100]")
		}
	case EventModeChange:
		if !slices.Contains([]string{"delta", "rate", "percent"}, c.Change) {
			return catalogInvalid(at+".change", "change is delta, rate or percent")
		}
	case EventModeAbsence:
		needField, needOp = false, false
	default:
		return catalogInvalid(at+".mode", "mode "+c.Mode+" is not a mode")
	}
	switch {
	case needField && c.Field == "":
		return catalogInvalid(at+".field", "a "+c.Mode+" condition names a field")
	case needOp && !slices.Contains(eventOps, c.Op):
		return catalogInvalid(at+".op", "operator "+c.Op+" is not gt, ge, lt, le, eq or ne")
	case !needOp && c.Op != "":
		return catalogInvalid(at+".op", "this "+c.Mode+" condition takes no operator")
	case needWindow && c.Window <= 0:
		return catalogInvalid(at+".window", "a "+c.Mode+" condition needs a positive window")
	case !needWindow && c.Window != 0:
		return catalogInvalid(at+".window", "a "+c.Mode+" condition takes no window")
	}
	return nil
}

func validateCatalogFields(path string, fs []CatalogField) error {
	seen := map[string]bool{}
	for i, f := range fs {
		p := fmt.Sprintf("%s[%d]", path, i)
		if f.Name == "" || seen[f.Name] {
			return catalogInvalid(p, "a field name is set and unique")
		}
		seen[f.Name] = true
		if err := validateCatalogField(p, &f); err != nil {
			return err
		}
	}
	return nil
}

func validateCatalogField(p string, f *CatalogField) error {
	if !slices.Contains(catalogFieldTypes, f.Type) {
		return catalogInvalid(p+".type", "unknown type "+f.Type)
	}
	if f.Min != nil && f.Max != nil && *f.Min > *f.Max {
		return catalogInvalid(p, "min is above max")
	}
	switch f.Type {
	case CatalogFieldEnum:
		if len(f.Values) == 0 {
			return catalogInvalid(p+".values", "an enum has values")
		}
	case CatalogFieldList:
		if f.Elem == nil {
			return catalogInvalid(p+".elem", "a list has an element type")
		}
		return validateCatalogField(p+".elem", f.Elem)
	case CatalogFieldObject:
		if len(f.Fields) == 0 {
			return catalogInvalid(p+".fields", "an object has fields")
		}
		return validateCatalogFields(p+".fields", f.Fields)
	}
	return nil
}

func validateAuditCatalog(a *AuditCatalog) error {
	if a == nil {
		return nil
	}
	types := map[string]bool{}
	for i, t := range a.EventTypes {
		p := fmt.Sprintf("audit.eventTypes[%d]", i)
		if t.Type == "" || strings.HasPrefix(t.Type, "asn.") || types[t.Type] {
			return catalogInvalid(p, "an event type is set, unique and does not start with asn.")
		}
		types[t.Type] = true
	}
	codes := map[int64]bool{}
	for i, o := range a.Operations {
		p := fmt.Sprintf("audit.operations[%d]", i)
		if codes[o.Code] {
			return catalogInvalid(p, "an operation code is unique")
		}
		if !types[o.EventType] {
			return catalogInvalid(p, "the event type "+o.EventType+" is not declared")
		}
		codes[o.Code] = true
	}
	objects := map[string]bool{}
	for i, o := range a.ObjectTypes {
		if o.Type == "" || objects[o.Type] {
			return catalogInvalid(fmt.Sprintf("audit.objectTypes[%d]", i), "an object type is set and unique")
		}
		objects[o.Type] = true
	}
	if len(a.TagKeys) > AuditTagKeysMax {
		return catalogInvalid("audit.tagKeys", fmt.Sprintf("at most %d tag keys", AuditTagKeysMax))
	}
	keys := map[string]bool{}
	for i, t := range a.TagKeys {
		p := fmt.Sprintf("audit.tagKeys[%d]", i)
		k := t.Key
		switch {
		case !auditTagKey.MatchString(k):
			return catalogInvalid(p, "a tag key is an upper-case letter followed by letters, digits and '_'")
		case slices.Contains(AuditReservedTagKeys, k):
			return catalogInvalid(p, k+" is set by ASN or the audit store")
		case keys[k]:
			return catalogInvalid(p, k+" is declared twice")
		case !hasName(t.Names):
			return catalogInvalid(p, "a tag key has a display name")
		case len(t.Values) > AuditTagValuesMax:
			return catalogInvalid(p, fmt.Sprintf("at most %d values", AuditTagValuesMax))
		}
		keys[k] = true
		values := map[string]bool{}
		for j, v := range t.Values {
			vp := fmt.Sprintf("%s.values[%d]", p, j)
			switch {
			case v.Value == "" || values[v.Value]:
				return catalogInvalid(vp, "a value is set and unique")
			case !hasName(v.Names):
				return catalogInvalid(vp, "a value has a display name")
			}
			values[v.Value] = true
		}
	}
	return nil
}

// hasName reports whether names holds a non-empty display name.
func hasName(names map[string]string) bool {
	for _, n := range names {
		if n != "" {
			return true
		}
	}
	return false
}

var auditTagKey = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)
