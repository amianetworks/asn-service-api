// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"encoding/json"
	"errors"
	"maps"
	"math"
	"strings"
	"testing"
)

// sampleTemplate uses every type and every widget once.
func sampleTemplate() *MetricsTemplate {
	texts := map[string]string{}
	for _, k := range []string{"t", "p", "f_total", "f_ok", "f_bad", "f_cpu", "f_state", "o_up", "o_down",
		"f_err", "f_top", "f_node", "f_flow", "f_up", "f_t", "f_v", "f_on", "f_since", "f_load", "h", "s"} {
		texts[k] = k
	}
	point := &MetricsField{Type: MetricsTypeObject, Fields: map[string]*MetricsField{
		"t": {Type: MetricsTypeTime, Name: "f_t"},
		"v": {Type: MetricsTypeFloat, Name: "f_v", Unit: MetricsUnitBitsPerSec},
	}}
	return &MetricsTemplate{
		Title: "t",
		Fields: map[string]*MetricsField{
			"total": {Type: MetricsTypeInt, Name: "f_total", Unit: MetricsUnitCount, Suffix: "s"},
			"ok":    {Type: MetricsTypeInt, Name: "f_ok", Unit: MetricsUnitCount, Level: MetricsLevelOK},
			"bad":   {Type: MetricsTypeInt, Name: "f_bad", Unit: MetricsUnitCount, Level: MetricsLevelHigh},
			"cpu":   {Type: MetricsTypeFloat, Name: "f_cpu", Unit: MetricsUnitPercent, Precision: new(0)},
			"load":  {Type: MetricsTypeFloat, Name: "f_load", Unit: MetricsUnitNone, Help: "h"},
			"on":    {Type: MetricsTypeBool, Name: "f_on", Scope: MetricsScopeService},
			"since": {Type: MetricsTypeTime, Name: "f_since"},
			"err":   {Type: MetricsTypeString, Name: "f_err"},
			"state": {Type: MetricsTypeEnum, Name: "f_state", Options: []*MetricsOption{
				{Key: "up", Name: "o_up", Level: MetricsLevelOK}, {Key: "down", Name: "o_down"}}},
			"top": {Type: MetricsTypeList, Name: "f_top", Item: &MetricsField{Type: MetricsTypeObject,
				Fields: map[string]*MetricsField{
					"node": {Type: MetricsTypeString, Name: "f_node"},
					"cpu":  {Type: MetricsTypeFloat, Name: "f_cpu", Unit: MetricsUnitPercent},
				}}},
			"flow": {Type: MetricsTypeObject, Name: "f_flow", Fields: map[string]*MetricsField{
				"up": {Type: MetricsTypeList, Name: "f_up", Item: point},
			}},
		},
		Layout: []*MetricsPanel{{Title: "p", Widgets: []*MetricsWidget{
			{Widget: MetricsWidgetStat, Field: "ok", Total: "total", Secondary: "bad", Thresholds: &MetricsThresholds{Warn: 1, Error: 5}},
			{Widget: MetricsWidgetGauge, Field: "cpu", Thresholds: &MetricsThresholds{Warn: 70, Error: 90}},
			{Widget: MetricsWidgetBreakdown, Style: MetricsStyleBadges, Fields: []string{"ok", "bad"}},
			{Widget: MetricsWidgetRanking, Field: "top", Label: "node", Value: "cpu", Limit: 5},
			{Widget: MetricsWidgetLine, Series: []*MetricsSeries{{Field: "flow.up", X: "t", Y: "v", Level: MetricsLevelLow}}},
			{Widget: MetricsWidgetKV, Fields: []string{"load", "on", "since", "err"}, Help: "h"},
			{Widget: MetricsWidgetStatus, Field: "state", Detail: "err"},
			{Widget: MetricsWidgetTable, Field: "top", Columns: []string{"node", "cpu"}},
		}}},
		Dict: map[string]map[string]string{MetricsLangZhCN: texts, MetricsLangEnUS: texts},
	}
}

func TestValidateMetricsTemplateAccepts(t *testing.T) {
	if err := ValidateMetricsTemplate(sampleTemplate()); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMetricsTemplate(&MetricsTemplate{}); err != nil {
		t.Fatalf("empty template: %v", err)
	}
}

func TestValidateMetricsTemplateRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*MetricsTemplate)
		want   string // substring of the error
	}{
		{"nil field", func(m *MetricsTemplate) { m.Fields["x"] = nil }, "field is nil"},
		{"bad name", func(m *MetricsTemplate) { m.Fields["Bad"] = m.Fields["total"] }, "does not match"},
		{"unknown type", func(m *MetricsTemplate) { m.Fields["total"].Type = "number" }, "unknown type"},
		{"missing name", func(m *MetricsTemplate) { m.Fields["total"].Name = "" }, "name is required"},
		{"missing unit", func(m *MetricsTemplate) { m.Fields["total"].Unit = "" }, "unit"},
		{"unknown unit", func(m *MetricsTemplate) { m.Fields["total"].Unit = "kb" }, "unit"},
		{"unit on string", func(m *MetricsTemplate) { m.Fields["err"].Unit = MetricsUnitCount }, "int and float only"},
		{"unknown level", func(m *MetricsTemplate) { m.Fields["ok"].Level = "warn" }, "unknown level"},
		{"precision on int", func(m *MetricsTemplate) { m.Fields["total"].Precision = new(1) }, "float only"},
		{"precision range", func(m *MetricsTemplate) { m.Fields["cpu"].Precision = new(7) }, "outside"},
		{"scope on member", func(m *MetricsTemplate) { m.Fields["flow"].Fields["up"].Scope = MetricsScopeService }, "top-level"},
		{"unknown scope", func(m *MetricsTemplate) { m.Fields["on"].Scope = "global" }, "unknown scope"},
		{"empty options", func(m *MetricsTemplate) { m.Fields["state"].Options = nil }, "needs options"},
		{"duplicate option", func(m *MetricsTemplate) { m.Fields["state"].Options[1].Key = "up" }, "duplicate option"},
		{"options on string", func(m *MetricsTemplate) { m.Fields["err"].Options = m.Fields["state"].Options }, "enum only"},
		{"empty object", func(m *MetricsTemplate) { m.Fields["flow"].Fields = nil }, "needs fields"},
		{"list without item", func(m *MetricsTemplate) { m.Fields["top"].Item = nil }, "needs item"},
		{"item on object", func(m *MetricsTemplate) { m.Fields["flow"].Item = m.Fields["top"].Item }, "list only"},
		{"named item", func(m *MetricsTemplate) { m.Fields["top"].Item.Name = "f_top" }, "list item"},
		{"depth", func(m *MetricsTemplate) {
			m.Fields["flow"].Fields["up"].Item.Fields["deep"] = &MetricsField{Type: MetricsTypeList, Name: "f_t",
				Item: &MetricsField{Type: MetricsTypeInt, Unit: MetricsUnitCount}}
		}, "nesting"},
		{"empty panel", func(m *MetricsTemplate) { m.Layout[0].Widgets = nil }, "needs widgets"},
		{"unknown widget", func(m *MetricsTemplate) { m.Layout[0].Widgets[0].Widget = "chart" }, "unknown widget"},
		{"inapplicable attribute", func(m *MetricsTemplate) { m.Layout[0].Widgets[1].Columns = []string{"x"} }, "does not apply"},
		{"unresolved", func(m *MetricsTemplate) { m.Layout[0].Widgets[0].Field = "missing" }, "does not resolve"},
		{"crosses list", func(m *MetricsTemplate) { m.Layout[0].Widgets[1].Field = "top.cpu" }, "crosses"},
		{"stat wrong type", func(m *MetricsTemplate) { m.Layout[0].Widgets[0].Field = "err" }, "not accepted"},
		{"stat total unit", func(m *MetricsTemplate) { m.Layout[0].Widgets[0].Total = "cpu" }, "differs"},
		{"warn above error", func(m *MetricsTemplate) { m.Layout[0].Widgets[0].Thresholds.Warn = 9 }, "warn is above"},
		{"gauge unit", func(m *MetricsTemplate) { m.Layout[0].Widgets[1].Field = "load" }, "percent"},
		{"breakdown count", func(m *MetricsTemplate) { m.Layout[0].Widgets[2].Fields = []string{"ok"} }, "at least 2"},
		{"breakdown style", func(m *MetricsTemplate) { m.Layout[0].Widgets[2].Style = "" }, "style"},
		{"breakdown units", func(m *MetricsTemplate) { m.Layout[0].Widgets[2].Fields = []string{"ok", "cpu"} }, "differs"},
		{"ranking label", func(m *MetricsTemplate) { m.Layout[0].Widgets[3].Label = "cpu" }, "not accepted"},
		{"limit range", func(m *MetricsTemplate) { m.Layout[0].Widgets[3].Limit = 1001 }, "outside"},
		{"line x", func(m *MetricsTemplate) { m.Layout[0].Widgets[4].Series[0].X = "v" }, "not accepted"},
		{"second series x", func(m *MetricsTemplate) {
			m.Layout[0].Widgets[4].Series = append(m.Layout[0].Widgets[4].Series,
				&MetricsSeries{Field: "top", X: "cpu", Y: "cpu"})
		}, "not accepted"},
		{"kv list", func(m *MetricsTemplate) { m.Layout[0].Widgets[5].Fields = []string{"top"} }, "not accepted"},
		{"status type", func(m *MetricsTemplate) { m.Layout[0].Widgets[6].Field = "err" }, "not accepted"},
		{"table columns", func(m *MetricsTemplate) { m.Layout[0].Widgets[7].Columns = nil }, "needs columns"},
		{"missing language", func(m *MetricsTemplate) { delete(m.Dict, MetricsLangEnUS) }, "en-US is missing"},
		{"missing key", func(m *MetricsTemplate) { delete(m.Dict[MetricsLangZhCN], "f_cpu") }, `"f_cpu" is missing`},
	}
	for _, c := range cases {
		m := sampleTemplate()
		if c.name == "missing key" { // the two languages share one map in sampleTemplate
			zh := map[string]string{}
			maps.Copy(zh, m.Dict[MetricsLangZhCN])
			m.Dict[MetricsLangZhCN] = zh
		}
		c.mutate(m)
		err := ValidateMetricsTemplate(m)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want %q", c.name, err, c.want)
		}
	}
	if err := ValidateMetricsTemplate(nil); err == nil {
		t.Error("nil template accepted")
	}
}

type flowPoint struct {
	T int64   `json:"t"`
	V float64 `json:"v"`
}

func TestValidateMetricsAccepts(t *testing.T) {
	var noTop []map[string]any // typed nil slice encodes as null
	data := map[string]any{
		"total": 12, "ok": int64(10), "bad": 1.0, "cpu": 37.25, "load": 0.5, "on": true,
		"since": int64(1761800000000), "err": "", "state": "up", "top": noTop,
		"flow": map[string]any{"up": []flowPoint{{T: 1761800000000, V: 1200}}},
	}
	tree, err := ValidateMetrics(sampleTemplate(), data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tree["top"]; ok {
		t.Error("null key kept")
	}
	if tree["total"] != json.Number("12") {
		t.Errorf("total = %#v", tree["total"])
	}

	empty, err := ValidateMetrics(sampleTemplate(), nil)
	if err != nil || len(empty) != 0 {
		t.Errorf("nil map: %v %v", empty, err)
	}
}

func TestValidateMetricsRejects(t *testing.T) {
	long := make([]map[string]any, MetricsMaxListLen+1)
	for i := range long {
		long[i] = map[string]any{"node": "n", "cpu": 1}
	}
	cases := []struct {
		name string
		data map[string]any
		path string
	}{
		{"undeclared top", map[string]any{"x": 1}, "x"},
		{"undeclared nested", map[string]any{"flow": map[string]any{"down": 1}}, "flow.down"},
		{"int non-integral", map[string]any{"total": 1.5}, "total"},
		{"int too large", map[string]any{"total": int64(1)<<53 + 1}, "total"},
		{"time non-integral", map[string]any{"since": 1.5}, "since"},
		{"string type", map[string]any{"err": 1}, "err"},
		{"bool type", map[string]any{"on": "yes"}, "on"},
		{"enum value", map[string]any{"state": "gone"}, "state"},
		{"list element type", map[string]any{"top": []any{map[string]any{"node": 1}}}, "top[0].node"},
		{"null list element", map[string]any{"top": []any{nil}}, "top[0]"},
		{"list length", map[string]any{"top": long}, "top"},
		{"first of several", map[string]any{"z": 1, "a": 1}, "a"},
	}
	for _, c := range cases {
		_, err := ValidateMetrics(sampleTemplate(), c.data)
		var ve *MetricsValidationError
		if !errors.As(err, &ve) || ve.Path != c.path {
			t.Errorf("%s: got %v, want path %q", c.name, err, c.path)
		}
	}

	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := ValidateMetrics(sampleTemplate(), map[string]any{"cpu": f})
		var ve *MetricsValidationError
		if !errors.As(err, &ve) || ve.Path != "" {
			t.Errorf("%v: got %v, want an encode error without path", f, err)
		}
	}

	big := map[string]any{"err": strings.Repeat("x", MetricsMaxDataBytes)}
	if _, err := ValidateMetrics(sampleTemplate(), big); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("oversized data: %v", err)
	}

	bad := sampleTemplate()
	bad.Fields["total"].Unit = ""
	if _, err := ValidateMetrics(bad, map[string]any{}); err == nil {
		t.Error("invalid template accepted")
	}
}
