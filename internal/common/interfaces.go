package common

// Logger interface for debug and info logging across all packages
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
}
