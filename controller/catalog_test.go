// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"errors"
	"testing"
)

func validCatalog() *ServiceCatalog {
	return &ServiceCatalog{
		Service:  "svc",
		Display:  CatalogDisplay{Name: map[string]string{"en-US": "Svc"}},
		Grouping: CatalogGroupingNone,
		Items: []CatalogItem{{Name: "svc.scan", Demand: CatalogDemandAlways, Node: true,
			Fields: []CatalogItemField{{Name: "found", Type: CatalogItemNumber}},
			Tags:   []CatalogItemField{{Name: "Kind", Type: CatalogItemEnum, Values: []string{"a", "b"}}}}},
		Audit: &AuditCatalog{EventTypes: []AuditEventType{{Type: "svc.scan"}},
			Operations: []AuditOperation{{Code: 1, EventType: "svc.scan"}}, TagKeys: []string{"Role", "Kind_2"}},
		DefaultRules: []EventRule{{Name: "found", Item: "svc.scan", Level: EventLevelHigh,
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
	for name, mutate := range map[string]func(c *ServiceCatalog){
		"nil display":       func(c *ServiceCatalog) { c.Display.Name = nil },
		"item prefix":       func(c *ServiceCatalog) { c.Items[0].Name = "other.scan" },
		"reserved tag":      func(c *ServiceCatalog) { c.Items[0].Tags[0].Name = "NodeID" },
		"number tag":        func(c *ServiceCatalog) { c.Items[0].Tags[0].Type = CatalogItemNumber },
		"asn event type":    func(c *ServiceCatalog) { c.Audit.EventTypes[0].Type = "asn.x" },
		"undeclared type":   func(c *ServiceCatalog) { c.Audit.Operations[0].EventType = "svc.other" },
		"audit tag asn":     func(c *ServiceCatalog) { c.Audit.TagKeys = []string{"Caller"} },
		"audit tag store":   func(c *ServiceCatalog) { c.Audit.TagKeys = []string{"Scope"} },
		"audit tag lower":   func(c *ServiceCatalog) { c.Audit.TagKeys = []string{"role"} },
		"audit tag chars":   func(c *ServiceCatalog) { c.Audit.TagKeys = []string{"Role-x"} },
		"audit tag empty":   func(c *ServiceCatalog) { c.Audit.TagKeys = []string{""} },
		"audit tag twice":   func(c *ServiceCatalog) { c.Audit.TagKeys = []string{"Role", "Role"} },
		"audit tag count":   func(c *ServiceCatalog) { c.Audit.TagKeys = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"} },
		"rule with root":    func(c *ServiceCatalog) { c.DefaultRules[0].RootID = "r" },
		"rule other ns":     func(c *ServiceCatalog) { c.DefaultRules[0].Service = "asn" },
		"rule no condition": func(c *ServiceCatalog) { c.DefaultRules[0].Conditions = nil },
		"rule item":         func(c *ServiceCatalog) { c.DefaultRules[0].Item = "svc.other" },
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
