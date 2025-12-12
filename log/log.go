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
	Panicf(string, ...interface{})
	Fatalf(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Tracef(string, ...interface{})
	Panic(...interface{})
	Fatal(...interface{})
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Trace(...interface{})
}
