// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

// Lock is a cluster-wide distributed lock obtained via ASNController.InitLocker().
// Controller side only.
type Lock interface {
	// Lock acquires the lock for the given key.
	// identifier distinguishes the holder; pass the same value to Unlock to prevent
	// accidental release by a different caller holding the same key.
	Lock(key, identifier string) error

	// Unlock releases the lock. Only succeeds if identifier matches the value passed to Lock.
	Unlock(key, identifier string) error
}
