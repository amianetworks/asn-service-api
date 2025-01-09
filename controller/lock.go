// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import "time"

type Lock interface {
	Lock(string, *LockOption) error
	Unlock(string, *LockOption) error
}

type LockOption struct {
	// optional
	identifier string
	waiting    time.Duration
	holding    time.Duration
}

func (l *LockOption) WithIdentifier(identifier string) {
	l.identifier = identifier
}

func (l *LockOption) WithWaitingTime(waiting time.Duration) {
	l.waiting = waiting
}

func (l *LockOption) WithHoldingTime(holding time.Duration) {
	l.holding = holding
}
