// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package subscription

import (
	"time"

	"asn.amiasys.com/asn-service-api/v26/subscription/apple"
	"asn.amiasys.com/asn-service-api/v26/subscription/google"
	"asn.amiasys.com/asn-service-api/v26/subscription/stripe"
)

// Instance is the subscription interface obtained via ASNController.GetSubscription.
// All methods are goroutine-safe. Call every Add* during ASNServiceController.Start()
// before using the other methods.
type Instance interface {
	GetNotificationChannel() <-chan string

	AddApple(envConfig *apple.EnvConfig, apiConfig *apple.APIConfig) error
	AddGoogle(envConfig *google.EnvConfig) error
	AddStripe(config *stripe.Config) error

	RedeemCredential(accountID, provider, credential string, needMigrate bool) error
	GetCredentialInfo(provider, credential string) (*CredentialInfo, error)
	RefreshAccount(accountID string)

	ListSubscriptionRecords(accountID string, activeOnly bool, from, to time.Time, page, num int) ([]*SubscriptionRecord, int, error)

	GetStripeCheckoutUrl(accountID string, planIDs []string, trialDays int, successUrl, cancelUrl string) (string, error)
	GetStripeBillingPortalUrl(accountID, returnUrl string) (string, error)
	GetStripeProductInfo(planID string) (*Product, error)
}
