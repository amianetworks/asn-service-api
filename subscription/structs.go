// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

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
