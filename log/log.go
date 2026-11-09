// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package log

type Logger struct {
	R ASNLogger // rlog | rtlog: Runtime log (default)
	A ASNLogger // alog | apilog: API text log (gRPC, RESTful API, CLI, etc.); structured API records: capi.StructuredLogAPI
	P ASNLogger // plog | perflog: Performance text log (performance optimization purpose; file only)
	E ASNLogger // elog | entitylog: Entity text log (access, change, create, delete, etc.); structured entity records: capi.StructuredLogAPI
}

// ASNLogger is the logger interface
type ASNLogger interface {
	Panicf(string, ...any)
	Fatalf(string, ...any)
	Debugf(string, ...any)
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
	Tracef(string, ...any)
	Panic(...any)
	Fatal(...any)
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Trace(...any)
}
