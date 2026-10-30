// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Service metrics templates.
//
// A service returns, per network, a MetricsTemplate (GetMetricsTemplate) that
// declares the keys of its metrics data and how they are displayed, and the
// data itself (GetMetrics). The framework checks both with
// ValidateMetricsTemplate and ValidateMetrics on every call; services call the
// same functions in their unit tests.
//
// Text shown to users is never inline: Name, Title, Help, Suffix and option
// names are keys of MetricsTemplate.Dict, which holds one map per language.
// zh-CN and en-US are required.

// MetricsType is the type of a metrics field.
type MetricsType string

const (
	MetricsTypeInt    MetricsType = "int"    // integral JSON number, |v| <= 2^53
	MetricsTypeFloat  MetricsType = "float"  // finite JSON number
	MetricsTypeString MetricsType = "string" // JSON string
	MetricsTypeBool   MetricsType = "bool"   // JSON boolean
	MetricsTypeEnum   MetricsType = "enum"   // JSON string, one of Options[].Key
	MetricsTypeTime   MetricsType = "time"   // integral JSON number, Unix milliseconds
	MetricsTypeObject MetricsType = "object" // JSON object, members declared in Fields
	MetricsTypeList   MetricsType = "list"   // JSON array, element declared in Item
)

// MetricsUnit is the base unit of an int or float field. Data always carries
// the base unit; scaling for display is the consumer's.
type MetricsUnit string

const (
	MetricsUnitCount         MetricsUnit = "count"
	MetricsUnitPercent       MetricsUnit = "percent" // 0–100 by convention, not range-checked
	MetricsUnitBytes         MetricsUnit = "bytes"
	MetricsUnitBytesPerSec   MetricsUnit = "bytes_per_sec"
	MetricsUnitBitsPerSec    MetricsUnit = "bits_per_sec"
	MetricsUnitPacketsPerSec MetricsUnit = "packets_per_sec"
	MetricsUnitMs            MetricsUnit = "ms"
	MetricsUnitNone          MetricsUnit = "none"
)

// MetricsLevel is the colour level of a value, a series, or an enum option.
// Empty means MetricsLevelNone.
type MetricsLevel string

const (
	MetricsLevelNone     MetricsLevel = "none"
	MetricsLevelOK       MetricsLevel = "ok"
	MetricsLevelLow      MetricsLevel = "low"
	MetricsLevelMedium   MetricsLevel = "medium"
	MetricsLevelHigh     MetricsLevel = "high"
	MetricsLevelCritical MetricsLevel = "critical"
)

// MetricsScope says what a top-level field covers. Empty means
// MetricsScopeNetwork.
type MetricsScope string

const (
	// MetricsScopeNetwork covers the network and all its subnetworks.
	MetricsScopeNetwork MetricsScope = "network"
	// MetricsScopeService does not follow the network subtree (a service-wide
	// total, or a count of entities that are not ASN nodes). It is shown to
	// every viewer of any network, marked as service-wide.
	MetricsScopeService MetricsScope = "service"
)

// MetricsWidgetKind is the kind of a widget.
type MetricsWidgetKind string

const (
	MetricsWidgetStat      MetricsWidgetKind = "stat"
	MetricsWidgetGauge     MetricsWidgetKind = "gauge"
	MetricsWidgetBreakdown MetricsWidgetKind = "breakdown"
	MetricsWidgetRanking   MetricsWidgetKind = "ranking"
	MetricsWidgetLine      MetricsWidgetKind = "line"
	MetricsWidgetKV        MetricsWidgetKind = "kv"
	MetricsWidgetStatus    MetricsWidgetKind = "status"
	MetricsWidgetTable     MetricsWidgetKind = "table"
)

// MetricsStyle is the style of a breakdown widget.
type MetricsStyle string

const (
	MetricsStylePie    MetricsStyle = "pie"
	MetricsStyleBars   MetricsStyle = "bars"
	MetricsStyleBadges MetricsStyle = "badges"
)

// Metrics language keys required in MetricsTemplate.Dict.
const (
	MetricsLangZhCN = "zh-CN"
	MetricsLangEnUS = "en-US"
)

// Metrics limits.
const (
	MetricsMaxTemplateBytes = 256 << 10 // encoded template
	MetricsMaxDataBytes     = 1 << 20   // encoded data
	MetricsMaxListLen       = 1000      // elements of one list
	MetricsMaxDepth         = 3         // nested object/list levels from the top
	MetricsMaxPrecision     = 6
	MetricsMaxLimit         = 1000
	metricsMaxInt           = 1 << 53
)

// MetricsTemplate declares a service's metrics for one network. An empty
// template (no fields, no layout) means the service reports no metrics there.
type MetricsTemplate struct {
	// Title is an optional dict key labelling the service; consumers use the
	// service name when it is unset.
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Fields declares every key the data may contain.
	Fields map[string]*MetricsField `json:"fields,omitempty" yaml:"fields,omitempty"`

	// Layout lists the panels in display order.
	Layout []*MetricsPanel `json:"layout,omitempty" yaml:"layout,omitempty"`

	// Dict maps language (MetricsLangZhCN, MetricsLangEnUS) to key to text.
	Dict map[string]map[string]string `json:"dict,omitempty" yaml:"dict,omitempty"`
}

// MetricsField declares one key of the data. Which attributes apply depends
// on Type; see ValidateMetricsTemplate.
type MetricsField struct {
	Type MetricsType `json:"type" yaml:"type"`

	// Name is a dict key; required except on a list Item.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Unit is required for int and float.
	Unit MetricsUnit `json:"unit,omitempty" yaml:"unit,omitempty"`

	// Precision is the decimal places of a float, 0–6; nil means the
	// consumer's default (2).
	Precision *int `json:"precision,omitempty" yaml:"precision,omitempty"`

	// Suffix is a dict key shown after an int or float value.
	Suffix string `json:"suffix,omitempty" yaml:"suffix,omitempty"`

	// Help is a dict key shown as a hint.
	Help string `json:"help,omitempty" yaml:"help,omitempty"`

	// Level colours an int or float value.
	Level MetricsLevel `json:"level,omitempty" yaml:"level,omitempty"`

	// Scope applies to top-level fields only.
	Scope MetricsScope `json:"scope,omitempty" yaml:"scope,omitempty"`

	// Options lists every value of an enum.
	Options []*MetricsOption `json:"options,omitempty" yaml:"options,omitempty"`

	// Fields declares the members of an object.
	Fields map[string]*MetricsField `json:"fields,omitempty" yaml:"fields,omitempty"`

	// Item declares the element of a list.
	Item *MetricsField `json:"item,omitempty" yaml:"item,omitempty"`
}

// MetricsOption is one value of an enum field.
type MetricsOption struct {
	Key   string       `json:"key" yaml:"key"`
	Name  string       `json:"name" yaml:"name"` // dict key
	Level MetricsLevel `json:"level,omitempty" yaml:"level,omitempty"`
}

// MetricsPanel groups widgets.
type MetricsPanel struct {
	Title   string           `json:"title,omitempty" yaml:"title,omitempty"` // dict key
	Widgets []*MetricsWidget `json:"widgets" yaml:"widgets"`
}

// MetricsWidget displays fields. References are member names joined by "."
// from the top of the data and never cross a list; Label, Value, Columns and
// series X / Y name members of a list element. Which attributes apply depends
// on Widget; see ValidateMetricsTemplate.
type MetricsWidget struct {
	Widget MetricsWidgetKind `json:"widget" yaml:"widget"`
	Title  string            `json:"title,omitempty" yaml:"title,omitempty"` // dict key
	Help   string            `json:"help,omitempty" yaml:"help,omitempty"`   // dict key

	Field      string             `json:"field,omitempty" yaml:"field,omitempty"`
	Fields     []string           `json:"fields,omitempty" yaml:"fields,omitempty"`
	Total      string             `json:"total,omitempty" yaml:"total,omitempty"`
	Secondary  string             `json:"secondary,omitempty" yaml:"secondary,omitempty"`
	Thresholds *MetricsThresholds `json:"thresholds,omitempty" yaml:"thresholds,omitempty"`
	Style      MetricsStyle       `json:"style,omitempty" yaml:"style,omitempty"`
	Label      string             `json:"label,omitempty" yaml:"label,omitempty"`
	Value      string             `json:"value,omitempty" yaml:"value,omitempty"`
	Limit      int                `json:"limit,omitempty" yaml:"limit,omitempty"` // 0: unset; else 1–1000
	Series     []*MetricsSeries   `json:"series,omitempty" yaml:"series,omitempty"`
	Detail     string             `json:"detail,omitempty" yaml:"detail,omitempty"`
	Columns    []string           `json:"columns,omitempty" yaml:"columns,omitempty"`
}

// MetricsSeries is one series of a line widget.
type MetricsSeries struct {
	Field string       `json:"field" yaml:"field"`                   // list of object
	X     string       `json:"x" yaml:"x"`                           // time member
	Y     string       `json:"y" yaml:"y"`                           // int or float member
	Name  string       `json:"name,omitempty" yaml:"name,omitempty"` // dict key
	Level MetricsLevel `json:"level,omitempty" yaml:"level,omitempty"`
}

// MetricsThresholds colours a value: below Warn ok, from Warn medium, from
// Error high. Higher is worse. Thresholds take precedence over a field level.
type MetricsThresholds struct {
	Warn  float64 `json:"warn" yaml:"warn"`
	Error float64 `json:"error" yaml:"error"`
}

// MetricsValidationError reports the first contract violation found.
type MetricsValidationError struct {
	// Path locates the violation: "." between members, "[i]" for list
	// elements, e.g. "flow.upload[3].v"; for a template, "fields.<name>…" or
	// "layout[i].widgets[j]…". Empty when no path applies.
	Path string
	// Rule names the broken rule.
	Rule string
}

func (e *MetricsValidationError) Error() string {
	if e.Path == "" {
		return "metrics: " + e.Rule
	}
	return "metrics: " + e.Path + ": " + e.Rule
}

func metricsErr(path, format string, args ...any) error {
	return &MetricsValidationError{Path: path, Rule: fmt.Sprintf(format, args...)}
}

var metricsNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidateMetricsTemplate checks a template against the metrics contract:
// vocabulary values and names, required and inapplicable attributes, option
// keys, ranges, nesting depth, size, widget references and their types and
// units, and dict coverage in zh-CN and en-US. Keys are checked in sorted
// order and the first violation is returned.
func ValidateMetricsTemplate(t *MetricsTemplate) error {
	if t == nil {
		return metricsErr("", "template is nil")
	}
	encoded, err := json.Marshal(t)
	if err != nil {
		return metricsErr("", "template does not encode: %v", err)
	}
	if len(encoded) > MetricsMaxTemplateBytes {
		return metricsErr("", "template exceeds %d bytes", MetricsMaxTemplateBytes)
	}

	v := &templateValidator{t: t, keys: map[string]string{}}
	if t.Title != "" {
		v.ref("title", t.Title)
	}
	for _, name := range sortedKeys(t.Fields) {
		if err = v.field("fields."+name, name, t.Fields[name], 0, fieldTop); err != nil {
			return err
		}
	}
	for i, p := range t.Layout {
		if err = v.panel(fmt.Sprintf("layout[%d]", i), p); err != nil {
			return err
		}
	}
	return v.dict()
}

type fieldPosition int

const (
	fieldTop fieldPosition = iota
	fieldMember
	fieldItem
)

type templateValidator struct {
	t    *MetricsTemplate
	keys map[string]string // dict key -> first path that references it
}

func (v *templateValidator) ref(path, key string) {
	if _, ok := v.keys[key]; !ok {
		v.keys[key] = path
	}
}

func (v *templateValidator) field(path, name string, f *MetricsField, depth int, pos fieldPosition) error {
	if f == nil {
		return metricsErr(path, "field is nil")
	}
	if pos != fieldItem && !metricsNameRE.MatchString(name) {
		return metricsErr(path, "name %q does not match %s", name, metricsNameRE)
	}
	numeric := f.Type == MetricsTypeInt || f.Type == MetricsTypeFloat
	switch f.Type {
	case MetricsTypeInt, MetricsTypeFloat, MetricsTypeString, MetricsTypeBool,
		MetricsTypeEnum, MetricsTypeTime, MetricsTypeObject, MetricsTypeList:
	default:
		return metricsErr(path, "unknown type %q", f.Type)
	}

	if pos == fieldItem {
		if f.Name != "" || f.Help != "" {
			return metricsErr(path, "name and help do not apply to a list item")
		}
	} else {
		if f.Name == "" {
			return metricsErr(path, "name is required")
		}
		v.ref(path+".name", f.Name)
		if f.Help != "" {
			v.ref(path+".help", f.Help)
		}
	}

	if numeric {
		if !validUnit(f.Unit) {
			return metricsErr(path, "unit %q is missing or unknown", f.Unit)
		}
		if f.Suffix != "" {
			v.ref(path+".suffix", f.Suffix)
		}
		if !validLevel(f.Level) {
			return metricsErr(path, "unknown level %q", f.Level)
		}
	} else if f.Unit != "" || f.Suffix != "" || f.Level != "" {
		return metricsErr(path, "unit, suffix and level apply to int and float only")
	}
	if f.Precision != nil {
		if f.Type != MetricsTypeFloat {
			return metricsErr(path, "precision applies to float only")
		}
		if *f.Precision < 0 || *f.Precision > MetricsMaxPrecision {
			return metricsErr(path, "precision %d is outside 0–%d", *f.Precision, MetricsMaxPrecision)
		}
	}
	if f.Scope != "" {
		if pos != fieldTop {
			return metricsErr(path, "scope applies to top-level fields only")
		}
		if f.Scope != MetricsScopeNetwork && f.Scope != MetricsScopeService {
			return metricsErr(path, "unknown scope %q", f.Scope)
		}
	}

	if f.Type == MetricsTypeEnum {
		if len(f.Options) == 0 {
			return metricsErr(path, "enum needs options")
		}
		seen := map[string]bool{}
		for i, o := range f.Options {
			op := fmt.Sprintf("%s.options[%d]", path, i)
			if o == nil || o.Key == "" || o.Name == "" {
				return metricsErr(op, "option needs key and name")
			}
			if seen[o.Key] {
				return metricsErr(op, "duplicate option key %q", o.Key)
			}
			seen[o.Key] = true
			if !validLevel(o.Level) {
				return metricsErr(op, "unknown level %q", o.Level)
			}
			v.ref(op+".name", o.Name)
		}
	} else if len(f.Options) > 0 {
		return metricsErr(path, "options apply to enum only")
	}

	composite := f.Type == MetricsTypeObject || f.Type == MetricsTypeList
	if composite {
		depth++
		if depth > MetricsMaxDepth {
			return metricsErr(path, "nesting exceeds %d levels", MetricsMaxDepth)
		}
	}
	if f.Type == MetricsTypeObject {
		if len(f.Fields) == 0 {
			return metricsErr(path, "object needs fields")
		}
		for _, member := range sortedKeys(f.Fields) {
			if err := v.field(path+"."+member, member, f.Fields[member], depth, fieldMember); err != nil {
				return err
			}
		}
	} else if len(f.Fields) > 0 {
		return metricsErr(path, "fields apply to object only")
	}
	if f.Type == MetricsTypeList {
		if f.Item == nil {
			return metricsErr(path, "list needs item")
		}
		return v.field(path+".item", "", f.Item, depth, fieldItem)
	} else if f.Item != nil {
		return metricsErr(path, "item applies to list only")
	}
	return nil
}

func (v *templateValidator) panel(path string, p *MetricsPanel) error {
	if p == nil {
		return metricsErr(path, "panel is nil")
	}
	if p.Title != "" {
		v.ref(path+".title", p.Title)
	}
	if len(p.Widgets) == 0 {
		return metricsErr(path, "panel needs widgets")
	}
	for i, w := range p.Widgets {
		if err := v.widget(fmt.Sprintf("%s.widgets[%d]", path, i), w); err != nil {
			return err
		}
	}
	return nil
}

// widgetAttrs lists the attributes each widget accepts besides Widget,
// Title and Help.
var widgetAttrs = map[MetricsWidgetKind]map[string]bool{
	MetricsWidgetStat:      {"field": true, "total": true, "secondary": true, "thresholds": true},
	MetricsWidgetGauge:     {"field": true, "thresholds": true},
	MetricsWidgetBreakdown: {"fields": true, "style": true},
	MetricsWidgetRanking:   {"field": true, "label": true, "value": true, "limit": true},
	MetricsWidgetLine:      {"series": true},
	MetricsWidgetKV:        {"fields": true},
	MetricsWidgetStatus:    {"field": true, "detail": true},
	MetricsWidgetTable:     {"field": true, "columns": true, "limit": true},
}

func (v *templateValidator) widget(path string, w *MetricsWidget) error {
	if w == nil {
		return metricsErr(path, "widget is nil")
	}
	allowed, ok := widgetAttrs[w.Widget]
	if !ok {
		return metricsErr(path, "unknown widget %q", w.Widget)
	}
	set := map[string]bool{
		"field": w.Field != "", "fields": len(w.Fields) > 0, "total": w.Total != "",
		"secondary": w.Secondary != "", "thresholds": w.Thresholds != nil, "style": w.Style != "",
		"label": w.Label != "", "value": w.Value != "", "limit": w.Limit != 0,
		"series": len(w.Series) > 0, "detail": w.Detail != "", "columns": len(w.Columns) > 0,
	}
	for _, attr := range sortedKeys(set) {
		if set[attr] && !allowed[attr] {
			return metricsErr(path, "%s does not apply to %s", attr, w.Widget)
		}
	}
	if w.Title != "" {
		v.ref(path+".title", w.Title)
	}
	if w.Help != "" {
		v.ref(path+".help", w.Help)
	}
	if w.Thresholds != nil && w.Thresholds.Warn > w.Thresholds.Error {
		return metricsErr(path, "thresholds warn is above error")
	}
	if w.Limit != 0 && (w.Limit < 1 || w.Limit > MetricsMaxLimit) {
		return metricsErr(path, "limit %d is outside 1–%d", w.Limit, MetricsMaxLimit)
	}

	switch w.Widget {
	case MetricsWidgetStat:
		f, err := v.numeric(path+".field", w.Field)
		if err != nil {
			return err
		}
		if w.Total != "" {
			total, err := v.numeric(path+".total", w.Total)
			if err != nil {
				return err
			}
			if total.Unit != f.Unit {
				return metricsErr(path+".total", "unit %q differs from field unit %q", total.Unit, f.Unit)
			}
		}
		if w.Secondary != "" {
			if _, err = v.numeric(path+".secondary", w.Secondary); err != nil {
				return err
			}
		}
	case MetricsWidgetGauge:
		f, err := v.numeric(path+".field", w.Field)
		if err != nil {
			return err
		}
		if f.Unit != MetricsUnitPercent {
			return metricsErr(path+".field", "gauge needs unit percent")
		}
	case MetricsWidgetBreakdown:
		if len(w.Fields) < 2 {
			return metricsErr(path, "breakdown needs at least 2 fields")
		}
		if w.Style != MetricsStylePie && w.Style != MetricsStyleBars && w.Style != MetricsStyleBadges {
			return metricsErr(path, "style %q is missing or unknown", w.Style)
		}
		if err := v.sameUnit(path, w.Fields); err != nil {
			return err
		}
	case MetricsWidgetRanking:
		elem, err := v.listOfObject(path+".field", w.Field)
		if err != nil {
			return err
		}
		if _, err = v.member(path+".label", elem, w.Label, MetricsTypeString); err != nil {
			return err
		}
		if _, err = v.member(path+".value", elem, w.Value, MetricsTypeInt, MetricsTypeFloat); err != nil {
			return err
		}
	case MetricsWidgetLine:
		if len(w.Series) == 0 {
			return metricsErr(path, "line needs series")
		}
		var unit MetricsUnit
		for i, s := range w.Series {
			sp := fmt.Sprintf("%s.series[%d]", path, i)
			if s == nil {
				return metricsErr(sp, "series is nil")
			}
			elem, err := v.listOfObject(sp+".field", s.Field)
			if err != nil {
				return err
			}
			if _, err = v.member(sp+".x", elem, s.X, MetricsTypeTime); err != nil {
				return err
			}
			y, err := v.member(sp+".y", elem, s.Y, MetricsTypeInt, MetricsTypeFloat)
			if err != nil {
				return err
			}
			if i == 0 {
				unit = y.Unit
			} else if y.Unit != unit {
				return metricsErr(sp+".y", "unit %q differs from %q", y.Unit, unit)
			}
			if !validLevel(s.Level) {
				return metricsErr(sp, "unknown level %q", s.Level)
			}
			if s.Name != "" {
				v.ref(sp+".name", s.Name)
			}
		}
	case MetricsWidgetKV:
		if len(w.Fields) == 0 {
			return metricsErr(path, "kv needs fields")
		}
		for i, ref := range w.Fields {
			if _, err := v.scalar(fmt.Sprintf("%s.fields[%d]", path, i), ref); err != nil {
				return err
			}
		}
	case MetricsWidgetStatus:
		if _, err := v.typed(path+".field", w.Field, MetricsTypeEnum); err != nil {
			return err
		}
		if w.Detail != "" {
			if _, err := v.typed(path+".detail", w.Detail, MetricsTypeString); err != nil {
				return err
			}
		}
	case MetricsWidgetTable:
		elem, err := v.listOfObject(path+".field", w.Field)
		if err != nil {
			return err
		}
		if len(w.Columns) == 0 {
			return metricsErr(path, "table needs columns")
		}
		for i, c := range w.Columns {
			if _, err = v.member(fmt.Sprintf("%s.columns[%d]", path, i), elem, c, scalarTypes...); err != nil {
				return err
			}
		}
	}
	return nil
}

var scalarTypes = []MetricsType{
	MetricsTypeInt, MetricsTypeFloat, MetricsTypeString, MetricsTypeBool, MetricsTypeEnum, MetricsTypeTime,
}

// resolve walks a dotted reference through object members, starting at
// fields; it never crosses a list.
func resolve(fields map[string]*MetricsField, ref string) (*MetricsField, error) {
	if ref == "" {
		return nil, fmt.Errorf("reference is empty")
	}
	parts := strings.Split(ref, ".")
	var f *MetricsField
	for i, part := range parts {
		if i > 0 {
			if f.Type != MetricsTypeObject {
				return nil, fmt.Errorf("reference %q crosses a %s", ref, f.Type)
			}
			fields = f.Fields
		}
		f = fields[part]
		if f == nil {
			return nil, fmt.Errorf("reference %q does not resolve", ref)
		}
	}
	return f, nil
}

func (v *templateValidator) typed(path, ref string, types ...MetricsType) (*MetricsField, error) {
	f, err := resolve(v.t.Fields, ref)
	if err != nil {
		return nil, metricsErr(path, "%v", err)
	}
	return f, checkType(path, f, types)
}

func (v *templateValidator) numeric(path, ref string) (*MetricsField, error) {
	return v.typed(path, ref, MetricsTypeInt, MetricsTypeFloat)
}

func (v *templateValidator) scalar(path, ref string) (*MetricsField, error) {
	return v.typed(path, ref, scalarTypes...)
}

func (v *templateValidator) sameUnit(path string, refs []string) error {
	var unit MetricsUnit
	for i, ref := range refs {
		fp := fmt.Sprintf("%s.fields[%d]", path, i)
		f, err := v.numeric(fp, ref)
		if err != nil {
			return err
		}
		if i == 0 {
			unit = f.Unit
		} else if f.Unit != unit {
			return metricsErr(fp, "unit %q differs from %q", f.Unit, unit)
		}
	}
	return nil
}

// listOfObject resolves ref to a list whose element is an object and returns
// the element.
func (v *templateValidator) listOfObject(path, ref string) (*MetricsField, error) {
	f, err := v.typed(path, ref, MetricsTypeList)
	if err != nil {
		return nil, err
	}
	if f.Item == nil || f.Item.Type != MetricsTypeObject {
		return nil, metricsErr(path, "list element is not an object")
	}
	return f.Item, nil
}

func (v *templateValidator) member(path string, elem *MetricsField, ref string, types ...MetricsType) (*MetricsField, error) {
	f, err := resolve(elem.Fields, ref)
	if err != nil {
		return nil, metricsErr(path, "%v", err)
	}
	return f, checkType(path, f, types)
}

func checkType(path string, f *MetricsField, types []MetricsType) error {
	if slices.Contains(types, f.Type) {
		return nil
	}
	return metricsErr(path, "type %s is not accepted here", f.Type)
}

func (v *templateValidator) dict() error {
	if len(v.keys) == 0 && len(v.t.Fields) == 0 && len(v.t.Layout) == 0 {
		return nil
	}
	for _, lang := range []string{MetricsLangZhCN, MetricsLangEnUS} {
		texts, ok := v.t.Dict[lang]
		if !ok {
			return metricsErr("dict", "language %s is missing", lang)
		}
		for _, key := range sortedKeys(v.keys) {
			if _, ok = texts[key]; !ok {
				return metricsErr(v.keys[key], "dict key %q is missing in %s", key, lang)
			}
		}
	}
	return nil
}

func validUnit(u MetricsUnit) bool {
	switch u {
	case MetricsUnitCount, MetricsUnitPercent, MetricsUnitBytes, MetricsUnitBytesPerSec,
		MetricsUnitBitsPerSec, MetricsUnitPacketsPerSec, MetricsUnitMs, MetricsUnitNone:
		return true
	}
	return false
}

func validLevel(l MetricsLevel) bool {
	switch l {
	case "", MetricsLevelNone, MetricsLevelOK, MetricsLevelLow, MetricsLevelMedium,
		MetricsLevelHigh, MetricsLevelCritical:
		return true
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ValidateMetrics checks the data a service returned from GetMetrics against
// t. It first validates t, then encodes data with encoding/json (NaN and ±Inf
// fail here, without a path) and decodes it with exact numbers (json.Number).
// A null value at any key means the key is absent and is removed; a nil data
// map is an empty result. Every key must be declared, every value must match
// its field's type, enum values must be options, and the size limits must
// hold. It returns the normalized tree.
func ValidateMetrics(t *MetricsTemplate, data map[string]any) (map[string]any, error) {
	if err := ValidateMetricsTemplate(t); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, metricsErr("", "data does not encode: %v", err)
	}
	if len(encoded) > MetricsMaxDataBytes {
		return nil, metricsErr("", "data exceeds %d bytes", MetricsMaxDataBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(encoded))
	dec.UseNumber()
	var tree map[string]any
	if err = dec.Decode(&tree); err != nil {
		return nil, metricsErr("", "data does not decode: %v", err)
	}
	if tree == nil {
		return map[string]any{}, nil
	}
	if err = checkObject("", tree, t.Fields); err != nil {
		return nil, err
	}
	return tree, nil
}

func checkObject(path string, obj map[string]any, fields map[string]*MetricsField) error {
	for _, key := range sortedKeys(obj) {
		kp := key
		if path != "" {
			kp = path + "." + key
		}
		f := fields[key]
		if f == nil {
			return metricsErr(kp, "key is not declared")
		}
		if obj[key] == nil {
			delete(obj, key)
			continue
		}
		if err := checkValue(kp, obj[key], f); err != nil {
			return err
		}
	}
	return nil
}

func checkValue(path string, value any, f *MetricsField) error {
	switch f.Type {
	case MetricsTypeInt, MetricsTypeTime:
		n, ok := value.(json.Number)
		if !ok {
			return metricsErr(path, "%s needs a number", f.Type)
		}
		i, err := strconv.ParseInt(n.String(), 10, 64)
		if err != nil || i > metricsMaxInt || i < -metricsMaxInt {
			return metricsErr(path, "%s needs an integer within ±2^53", f.Type)
		}
	case MetricsTypeFloat:
		n, ok := value.(json.Number)
		if !ok {
			return metricsErr(path, "float needs a number")
		}
		if _, err := n.Float64(); err != nil {
			return metricsErr(path, "float is out of range")
		}
	case MetricsTypeString:
		if _, ok := value.(string); !ok {
			return metricsErr(path, "string needs a string")
		}
	case MetricsTypeBool:
		if _, ok := value.(bool); !ok {
			return metricsErr(path, "bool needs a boolean")
		}
	case MetricsTypeEnum:
		s, ok := value.(string)
		if !ok {
			return metricsErr(path, "enum needs a string")
		}
		for _, o := range f.Options {
			if o.Key == s {
				return nil
			}
		}
		return metricsErr(path, "value is not an option")
	case MetricsTypeObject:
		obj, ok := value.(map[string]any)
		if !ok {
			return metricsErr(path, "object needs an object")
		}
		return checkObject(path, obj, f.Fields)
	case MetricsTypeList:
		list, ok := value.([]any)
		if !ok {
			return metricsErr(path, "list needs an array")
		}
		if len(list) > MetricsMaxListLen {
			return metricsErr(path, "list exceeds %d elements", MetricsMaxListLen)
		}
		for i, elem := range list {
			ep := fmt.Sprintf("%s[%d]", path, i)
			if elem == nil {
				return metricsErr(ep, "list element is null")
			}
			if err := checkValue(ep, elem, f.Item); err != nil {
				return err
			}
		}
	}
	return nil
}
