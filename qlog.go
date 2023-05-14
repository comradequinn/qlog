package qlog

import (
	"context"
	"io"
)

var defaultLog = New(OutputMaskAll, true)

// Sets the Writer used by the default logger
// This operation is intended for configuration during start-up. It is not safe for concurrent use.
func SetWriter(w io.Writer) {
	defaultLog.Writer = w
}

// Sets the outputmask used by the default logger
// This operation is intended for configuration during start-up. It is not safe for concurrent use.
func SetOutputMask(m int) {
	defaultLog.outputMask = m
}

// Sets whether the output of the default logger is JSON or logfmt
// This operation is intended for configuration during start-up. It is not safe for concurrent use.
// If this operation should be called before any call to SetLabels. If it is called after, those previously labels will be discarded
func SetOutputJSON(v bool) {
	l := New(defaultLog.outputMask, v)
	l.Writer = defaultLog.Writer
	defaultLog = l
}

// Sets labels to be included all logs written by the default logger
// This operation is intended for configuration during start-up. It is not safe for concurrent use.
func SetLabels(labels ...any) {
	l := New(defaultLog.outputMask, defaultLog.outputJSON, labels...)
	l.Writer = defaultLog.Writer
	defaultLog = l
}

// Writes a log with fatal severity to the default log and terminates the process
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Fatal(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Fatal(ctx context.Context, message string, err error, labels ...any) {
	defaultLog.Fatal(ctx, message, err, labels...)
}

// Writes a log with error severity to the default log
// Error is reserved for events that causes a primary task to fail, such a http request or queued task handler
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Error(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Error(ctx context.Context, message string, err error, labels ...any) {
	defaultLog.Error(ctx, message, err, labels...)
}

// Writes a log with warning severity to the default log
// Warning is reserved for events that cause a process or task to run sub-optimally but not fail, such downstream API calls that fail
// initially but are retried and subsequently succeed
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Warning(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Warning(ctx context.Context, message string, err error, labels ...any) {
	defaultLog.Warning(ctx, message, err, labels...)
}

// Writes a log with notice severity to the default log
// Notice is reserved for events that are expected but important, such as sytem start-up or shut-down
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Notice(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Notice(ctx context.Context, message string, labels ...any) {
	defaultLog.Notice(ctx, message, labels...)
}

// Writes a log with info severity to the default log
// Info is reserved for emitting high level detail about a process's internal activity, such as completing a http request or queued task
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Info(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Info(ctx context.Context, message string, labels ...any) {
	defaultLog.Info(ctx, message, labels...)
}

// Writes a log with debug severity and a label of trace=true to the default log
// trace is reserved for emitting data about IO within a process's internal activity, such as the content of requests received or generated
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Trace(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Trace(ctx context.Context, message string, labels ...any) {
	defaultLog.Trace(ctx, message, labels...)
}

// Writes a log with debug severity to the default log
// Debug is reserved for emitting low level detail about a process's internal activity, such as config data or current variable states
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	qlog.Debug(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func Debug(ctx context.Context, message string, labels ...any) {
	defaultLog.Debug(ctx, message, labels...)
}
