package applog

// Noop is a Logger that discards every message. Use it as a non-nil default
// (e.g. when a constructor receives a nil logger) so downstream code can call
// the logger directly without repeating `if logger != nil` guards.
var Noop Logger = noopLogger{}

type noopLogger struct{}

func (noopLogger) Debug(string, ...string) {
	// Intentionally discard debug messages.
}

func (noopLogger) Info(string, ...string) {
	// Intentionally discard informational messages.
}

func (noopLogger) Warn(string, ...string) {
	// Intentionally discard warning messages.
}

func (noopLogger) Error(string, ...string) {
	// Intentionally discard error messages.
}
