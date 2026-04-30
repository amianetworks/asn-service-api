// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package subscription

import (
	"net/http"

	"asn.amiasys.com/asn-service-api/v26/subscription/apple"
	"asn.amiasys.com/asn-service-api/v26/subscription/google"
	"asn.amiasys.com/asn-service-api/v26/subscription/stripe"
)

// Instance is the subscription interface obtained via ASNController.GetSubscription().
// All methods are goroutine-safe.
//
// Registration pattern: call Add* during ASNServiceController.Start(). Each returns:
//   - an HTTP webhook handler — mount it via the http.Handler returned from WebHandler()
//   - an errChan for async backend failures — monitor in a background goroutine
type Instance interface {
	// GetNotificationChannel returns a unified channel that emits a string token on any
	// subscription lifecycle event (new, renewal, cancellation) from any registered platform.
	GetNotificationChannel() <-chan string

	// AddApple registers the Apple App Store platform.
	// Returns a webhook handler to mount via WebHandler() and an errChan for async backend errors.
	// Call during Start().
	AddApple(envConfig *apple.EnvConfig, apiConfig *apple.APIConfig) (
		func(w http.ResponseWriter, r *http.Request), <-chan error, error)

	// RestoreApplePurchaseToken re-links an existing App Store purchase to an account.
	RestoreApplePurchaseToken(accountID, purchaseToken string) error

	// AddGoogle registers the Google Play platform.
	// Returns a webhook handler to mount via WebHandler() and an errChan for async backend errors.
	// Call during Start().
	AddGoogle(envConfig *google.EnvConfig, replayConfig *google.ReplayConfig) (
		func(w http.ResponseWriter, r *http.Request), <-chan error, error)

	// RestoreGooglePurchaseToken re-links an existing Play Store purchase to an account.
	RestoreGooglePurchaseToken(accountID, purchaseToken string) error

	// AddStripe registers the Stripe platform.
	// Returns a webhook handler to mount via WebHandler() and an errChan for async backend errors.
	// Call during Start().
	AddStripe(config *stripe.Config) (
		func(w http.ResponseWriter, r *http.Request), <-chan error, error)

	// GetStripeProductInfo returns product details (name, description, prices by currency) for a Stripe price ID.
	GetStripeProductInfo(priceID string) (*Product, error)

	// GetStripePaymentLink returns a Stripe Checkout URL for the given account, price, quantity, and redirect URL.
	GetStripePaymentLink(accountID, priceID string, quantity uint, redirectUrl string) (string, error)

	// GetStripeBillingPortalUrl returns a Stripe Customer Portal URL for the given account and return URL.
	GetStripeBillingPortalUrl(accountID, returnUrl string) (string, error)

	// GetUserSubscription returns the active subscription for the account, or (nil, false, nil) if none exists.
	GetUserSubscription(accountID string) (*Subscription, bool, error)

	// ListUserSubscriptions returns all active subscriptions across all accounts.
	ListUserSubscriptions() ([]*Subscription, error)

	// DeleteUserSubscription removes the local subscription record for the account.
	// Does not cancel the subscription on the platform side.
	DeleteUserSubscription(accountID string) (bool, error)
}
