// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"errors"
	"fmt"
	"testing"
)

func zh(name string) map[string]string { return map[string]string{"zh-CN": name} }

func tagKey(key string, values ...AuditTagValue) AuditTagKey {
	return AuditTagKey{Key: key, Names: zh(key), Values: values}
}

func tagKeys(keys ...string) []AuditTagKey {
	out := make([]AuditTagKey, len(keys))
	for i, k := range keys {
		out[i] = tagKey(k)
	}
	return out
}

// withLicense adds an item that is not node data.
func withLicense(c *ServiceCatalog) *ServiceCatalog {
	c.Items = append(c.Items, CatalogItem{Name: "license", Demand: CatalogDemandAlways,
		Fields: []CatalogItemField{{Name: "usage", Type: CatalogItemNumber}}})
	return c
}

func validCatalog() *ServiceCatalog {
	return &ServiceCatalog{
		Display:  CatalogDisplay{Name: map[string]string{"en-US": "Svc"}},
		Grouping: CatalogGroupingNone,
		Items: []CatalogItem{{Name: "scan", Demand: CatalogDemandAlways, Node: true,
			Fields: []CatalogItemField{{Name: "found", Type: CatalogItemNumber}},
			Tags:   []CatalogItemField{{Name: "Kind", Type: CatalogItemEnum, Values: []string{"a", "b"}}}}},
		Audit: &AuditCatalog{EventTypes: []AuditEventType{{Type: "svc.scan"}},
			Operations: []AuditOperation{{Code: 1, EventType: "svc.scan"}}, TagKeys: []AuditTagKey{tagKey("Role", AuditTagValue{Value: "admin", Names: zh("管理员")}), tagKey("Kind_2")}},
		DefaultRules: []EventRule{{Name: "found", Item: "scan", Level: EventLevelHigh,
			Conditions: []EventCondition{{Mode: EventModeThreshold, Field: "found", Op: "gt", Value: 0}}}},
	}
}

func TestValidateServiceCatalog(t *testing.T) {
	if err := ValidateServiceCatalog(validCatalog()); err != nil {
		t.Fatal(err)
	}
	ok := validCatalog()
	ok.DefaultRules[0].Conditions[0] = EventCondition{Mode: EventModeCount, Tag: "Kind", Op: "eq", Values: []string{"a"}, Window: 60e9, Count: 3}
	if err := ValidateServiceCatalog(ok); err != nil {
		t.Fatalf("categorical count: %v", err)
	}
	for _, scope := range []string{EventScopeRoot, EventScopePlatform} {
		ok := withLicense(validCatalog())
		ok.DefaultRules[0].Scope = scope
		ok.DefaultRules[0].Item, ok.DefaultRules[0].Conditions[0].Field = "license", "usage"
		if err := ValidateServiceCatalog(ok); err != nil {
			t.Fatalf("default rule scope %s: %v", scope, err)
		}
	}
	for name, mutate := range map[string]func(c *ServiceCatalog){
		"nil display":       func(c *ServiceCatalog) { c.Display.Name = nil },
		"item qualified":    func(c *ServiceCatalog) { c.Items[0].Name = "svc.scan" },
		"item empty":        func(c *ServiceCatalog) { c.Items[0].Name = "" },
		"item twice":        func(c *ServiceCatalog) { c.Items = append(c.Items, c.Items[0]) },
		"reserved tag":      func(c *ServiceCatalog) { c.Items[0].Tags[0].Name = "NodeID" },
		"number tag":        func(c *ServiceCatalog) { c.Items[0].Tags[0].Type = CatalogItemNumber },
		"asn event type":    func(c *ServiceCatalog) { c.Audit.EventTypes[0].Type = "asn.x" },
		"undeclared type":   func(c *ServiceCatalog) { c.Audit.Operations[0].EventType = "svc.other" },
		"audit tag asn":     func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("Caller") },
		"audit tag store":   func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("Scope") },
		"audit tag lower":   func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("role") },
		"audit tag chars":   func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("Role-x") },
		"audit tag empty":   func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("") },
		"audit tag twice":   func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("Role", "Role") },
		"audit tag count":   func(c *ServiceCatalog) { c.Audit.TagKeys = tagKeys("A", "B", "C", "D", "E", "F", "G", "H", "I") },
		"audit tag no name": func(c *ServiceCatalog) { c.Audit.TagKeys[0].Names = map[string]string{"zh-CN": ""} },
		"audit value empty": func(c *ServiceCatalog) { c.Audit.TagKeys[0].Values[0].Value = "" },
		"audit value twice": func(c *ServiceCatalog) {
			c.Audit.TagKeys[0].Values = append(c.Audit.TagKeys[0].Values, c.Audit.TagKeys[0].Values[0])
		},
		"audit value no name": func(c *ServiceCatalog) { c.Audit.TagKeys[0].Values[0].Names = nil },
		"audit value count": func(c *ServiceCatalog) {
			c.Audit.TagKeys[0].Values = nil
			for i := 0; i <= AuditTagValuesMax; i++ {
				c.Audit.TagKeys[0].Values = append(c.Audit.TagKeys[0].Values, AuditTagValue{Value: fmt.Sprint(i), Names: zh("值")})
			}
		},
		"rule with root":     func(c *ServiceCatalog) { c.DefaultRules[0].RootID = "r" },
		"rule all roots":     func(c *ServiceCatalog) { c.DefaultRules[0].Scope = EventScopeAllRoots },
		"rule platform node": func(c *ServiceCatalog) { c.DefaultRules[0].Scope = EventScopePlatform },
		"rule platform cond node": func(c *ServiceCatalog) {
			*c = *withLicense(c)
			r := &c.DefaultRules[0]
			r.Scope, r.Item, r.Conditions[0].Field = EventScopePlatform, "license", "usage"
			r.Conditions = append(r.Conditions, EventCondition{Item: "scan", Mode: EventModeThreshold, Field: "found", Op: "gt", Value: 0})
		},
		"rule other ns":     func(c *ServiceCatalog) { c.DefaultRules[0].Service = "asn" },
		"rule own ns named": func(c *ServiceCatalog) { c.DefaultRules[0].Service = "svc" },
		"rule qualified":    func(c *ServiceCatalog) { c.DefaultRules[0].Item = "svc.scan" },
		"rule no condition": func(c *ServiceCatalog) { c.DefaultRules[0].Conditions = nil },
		"rule item":         func(c *ServiceCatalog) { c.DefaultRules[0].Item = "other" },
		"rule level":        func(c *ServiceCatalog) { c.DefaultRules[0].Level = "severe" },
		"rule handling":     func(c *ServiceCatalog) { c.DefaultRules[0].Handling = "later" },
		"rule field":        func(c *ServiceCatalog) { c.DefaultRules[0].Conditions[0].Field = "lost" },
		"rule op":           func(c *ServiceCatalog) { c.DefaultRules[0].Conditions[0].Op = "gte" },
		"rule mode":         func(c *ServiceCatalog) { c.DefaultRules[0].Conditions[0].Mode = "sometimes" },
		"rule window":       func(c *ServiceCatalog) { c.DefaultRules[0].Conditions[0].Window = 1 },
		"rule values op": func(c *ServiceCatalog) {
			c.DefaultRules[0].Conditions[0] = EventCondition{Mode: EventModeInstant, Tag: "Kind", Values: []string{"a"}}
		},
		"rule tag": func(c *ServiceCatalog) {
			c.DefaultRules[0].Conditions[0] = EventCondition{Mode: EventModeInstant, Tag: "Lost", Op: "eq", Values: []string{"a"}}
		},
		"rule count": func(c *ServiceCatalog) {
			c.DefaultRules[0].Conditions[0] = EventCondition{Mode: EventModeCount, Window: 60e9}
		},
	} {
		c := validCatalog()
		mutate(c)
		var ce *CatalogError
		if err := ValidateServiceCatalog(c); !errors.As(err, &ce) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
