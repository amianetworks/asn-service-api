// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package subscription

import (
	"time"
)

type Method string

const (
	MethodAppStore   = "AppStore"
	MethodGooglePlay = "GooglePlay"
	MethodStripe     = "Stripe"
)

type Subscription struct {
	AccountID string // UUID

	SubscriptionMethod Method
	ProductID          string // Apple: transaction.productId; Google: winning lineItem.productId

	SubscriptionID   string
	SubscriptionTime time.Time

	StartTime time.Time
	EndTime   time.Time
}

type Currency string

type Product struct {
	Name        string
	Description string

	DefaultCurrency Currency
	DefaultPrice    int64

	PriceOptions map[Currency]int64
}

// EntitlementStatus describes whether the subscription behind a restore
// request still grants access.
type EntitlementStatus string

const (
	EntitlementActive  EntitlementStatus = "active"
	EntitlementExpired EntitlementStatus = "expired"
	EntitlementRevoked EntitlementStatus = "revoked"
	EntitlementUnknown EntitlementStatus = "unknown"
)

// OwnerRelation describes how the subscription is currently bound relative to
// the account asking to restore it.
type OwnerRelation string

const (
	OwnerCurrentUser OwnerRelation = "currentUser" // already bound to the requesting account
	OwnerOtherUser   OwnerRelation = "otherUser"   // bound to a different account
	OwnerUnbound     OwnerRelation = "unbound"     // claimable / not bound to any account
	OwnerUnknown     OwnerRelation = "unknown"     // backend has not ingested this purchase yet
)

// RestoreAction is the recommended next step for the client, given what a real
// restore would actually do on this provider.
type RestoreAction string

const (
	ActionSync               RestoreAction = "sync"               // already bound to the current account & valid; no migration needed
	ActionMigrate            RestoreAction = "migrate"            // bound to another account; restore will move it (needs user confirmation)
	ActionBind               RestoreAction = "bind"               // unbound/orphan; restore will bind it to the current account
	ActionReject             RestoreAction = "reject"             // bound to another account and this provider cannot migrate it here
	ActionRepurchaseRequired RestoreAction = "repurchaseRequired" // no valid entitlement (expired/revoked)
	ActionWait               RestoreAction = "wait"               // backend has not ingested this purchase yet; retry later
)

// RestorePreview is the read-only result of a restore precheck. It NEVER
// modifies any binding; it only reports what a real restore would do, so the
// client can decide whether to prompt the user before committing.
type RestorePreview struct {
	EntitlementStatus EntitlementStatus `json:"entitlementStatus"`
	OwnerRelation     OwnerRelation     `json:"ownerRelation"`
	OwnerID           string            `json:"ownerId,omitempty"`
	Action            RestoreAction     `json:"action"`

	ProductID string     `json:"productId,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}
