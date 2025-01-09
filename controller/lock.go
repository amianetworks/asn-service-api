// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

type Lock interface {
	Lock(key, identifier string) error
	Unlock(key, identifier string) error
}
