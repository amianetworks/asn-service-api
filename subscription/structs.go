// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package subscription

import (
	"time"
)

const (
	ProviderAppStore   = "AppStore"
	ProviderGooglePlay = "GooglePlay"
	ProviderStripe     = "Stripe"
)

const (
	SubStateActive       = "active"
	SubStateGrace        = "grace"
	SubStateBillingRetry = "billing_retry"
	SubStatePaused       = "paused"
	SubStateExpired      = "expired"
	SubStateRevoked      = "revoked"
)

const (
	RecordStateNormal   = "normal"
	RecordStateExtended = "extended"
	RecordStateReplaced = "replaced"
	RecordStateRefunded = "refunded"
	RecordStateRevoked  = "revoked"
)

type Subscription struct {
	SubscriptionID   string
	Provider         string
	ExternalID       string
	AccountID        string
	State            string
	BundleRef        string
	StripeCustomerID string

	CreatedAt time.Time
}

type SubscriptionRecord struct {
	ID             string
	SubscriptionID string
	AccountID      string
	PlanID         string

	StartTime     time.Time
	EndTime       time.Time
	ActualEndTime time.Time
	State         string

	PaymentID    string
	Amount       int64
	Currency     string
	RefundAmount int64
}

type CredentialInfo struct {
	State   string
	OwnerID string

	PlanID    string
	ExpiresAt time.Time
}

type Product struct {
	Name        string
	Description string

	DefaultCurrency string
	DefaultPrice    int64

	PriceOptions map[string]int64
}
