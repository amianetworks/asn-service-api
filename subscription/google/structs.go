// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package google

type EnvConfig struct {
	PackageName        string
	ServiceAccountJSON []byte
}

type ReplayConfig struct {
	// GCP project and the RTDN subscription id you own.
	ProjectID      string
	SubscriptionID string // e.g. "rtdn-sub"
}
