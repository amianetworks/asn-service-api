// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package apple

type EnvConfig struct {
	BundleID string
	Env      string
}

// APIConfig holds credentials to call App Store Server API.
// See: "Generating JSON Web Tokens for API requests".
type APIConfig struct {
	IssuerID      string // iss
	KeyID         string // kid
	PrivateKeyPEM []byte // .p8 content
}
