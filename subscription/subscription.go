// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package subscription

import (
	"net/http"

	"asn.amiasys.com/asn-service-api/v25/subscription/apple"
	"asn.amiasys.com/asn-service-api/v25/subscription/google"
	"asn.amiasys.com/asn-service-api/v25/subscription/stripe"
)

type Instance interface {
	AddApple(envConfig *apple.EnvConfig, apiConfig *apple.APIConfig) (
		func(w http.ResponseWriter, r *http.Request), <-chan error, error)
	AddGoogle(envConfig *google.EnvConfig, replayConfig *google.ReplayConfig) (
		func(w http.ResponseWriter, r *http.Request), <-chan error, error)
	AddStripe(config *stripe.Config) (
		func(w http.ResponseWriter, r *http.Request), <-chan error, error)
	GetStripePaymentLink(username, priceID string, quantity uint, redirectUrl string) (string, error)

	GetUserSubscription(accountID string) (*Subscription, bool, error)
	ListUserSubscriptions() ([]*Subscription, error)
	DeleteUserSubscription(accountID string) (bool, error)
}
