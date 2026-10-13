// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package log

type Logger struct {
	R ASNLogger // rlog | rtlog: Runtime log (default)
	A ASNLogger // alog | apilog: API logger (gRPC, RESTful API, CLI, etc.)
	P ASNLogger // plog | perflog: Performance logger (performance optimization purpose)
	E ASNLogger // elog | entitylog: Entity logger (access, change, create, delete, etc.)
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
