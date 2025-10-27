// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package stripe

// Config for Stripe provider.
type Config struct {
	APIKey        string // Stripe secret key, e.g. "sk_live_…"
	SigningSecret string // Webhook endpoint's signing secret, e.g. "whsec_…"
	// Optional HTTP client on stripe-go if you need a custom transport/timeouts.
}
