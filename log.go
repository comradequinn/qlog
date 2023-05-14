// Package qlog provides simple, fast, structured logging with built-in request/task tracing.
//
// Logs written with the same context are assigned the same, unique Trace ID which is written
// to each log in the form "trace=abc123def456". This allows related logs to be easily collated.
//
// Log attribute keys should be strings, but values may be of any type or expressed as a func() T.
// In the latter case, the func() T will not be evaluated unless the log is written. This can prevent
// expensive evaluations being performed for logs of a severity that is not enabled for output.
//
// Log output verbosity is controlled by configuring the OutputMask; either with the individual OutputFlags
// required, or by using one of the preset OutputMasks
package qlog

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	// Log is a individual Log instance carrying its own specific configuration
	Log struct {
		commonLabels string
		outputMask   int
		outputJSON   bool
		Writer       io.Writer
	}
	unexportedKey struct{}
)

// OutputMask flag for configuring output verbosity
const (
	OutputFlagNone    = 0b00000000
	OutputFlagFatal   = 0b00000001
	OutputFlagError   = 0b00000010
	OutputFlagWarning = 0b00000100
	OutputFlagNotice  = 0b00001000
	OutputFlagInfo    = 0b00010000
	OutputFlagTrace   = 0b00100000
	OutputFlagDebug   = 0b01000000
)

// Predefined OutputMask for configuring output verbosity
const (
	OutputMaskImportant = OutputFlagFatal | OutputFlagError | OutputFlagWarning | OutputFlagNotice
	OutputMaskDetail    = OutputMaskImportant | OutputFlagInfo
	OutputMaskAll       = OutputMaskDetail | OutputFlagDebug
)

// Exported configuration fields
var (
	// TimestampFormat defines the format that will be used timestamps in logs
	TimestampFormat = time.RFC3339Nano
	// TraceIDFieldName defines the key assigned to the Trace-ID in the log
	//
	// By default it is, `trace`; override this, if required, to align with
	//  conventions or tooling that supports a similar feature by uses a different field name
	TraceIDFieldName = "trace"
	// TraceID returns the Trace-ID associated with the passed ctx.
	// This allows it to be passed across process boundaries, for example
	// as a HTTP Header in a downstream API call
	//
	// By default it returns the unique key assigned to the ctx by the conventional
	// call to qlog.ContextFrom(); override this, if required, this to read a diffferent value
	// written by existing conventions or tooling that supports a similar feature
	TraceID = func(ctx context.Context) string {
		traceID, _ := ctx.Value(traceIDKey).(string)

		return traceID
	}
	// FatalFunc defines the function called by Fatal after writing the log
	//
	// By default this is `os.Exit(1)`; override this is different behavior is required
	FatalFunc = func() { os.Exit(1) }
)

var (
	mx         = sync.Mutex{} // outside of testing, all loggers are likely to be writing to the same destination (stderr), so they all share the same write lock
	timeNow    = time.Now
	traceIDKey = unexportedKey{}
	newSpanID  = func() func() string {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		pad := "XXXXXXXXXXXXXXXXXXX" // padding is to keep Trace-IDs the same length

		return func() string {
			return (strconv.Itoa(r.Int()) + pad)[:len(pad)]
		}
	}()
)

// ContextFrom creates a new context.Context with the passed Trace-IDs or a new, unique Trace-ID
// if traceID is an empty string
//
// This will cause logs generated from method calls that are passed the returned
// context.Context to share a common Trace-ID field value in the log output
func ContextFrom(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		panic("nil context passed to context-from")
	}

	if traceID == "" {
		traceID = newSpanID()
	}

	return context.WithValue(ctx, traceIDKey, traceID)
}

// New creates a new Log with the specified output verbosity, common labels and
// whether JSON or logfmt output is required
func New(outputMask int, outputJSON bool, labels ...any) *Log {
	sb := strings.Builder{}
	sb.Grow(1000)

	writeLabels(&sb, outputJSON, labels)

	return &Log{outputMask: outputMask, outputJSON: outputJSON, commonLabels: sb.String(), Writer: os.Stderr}
}

// WithLabels creates a new Log with the same labels as the receiver Log
// but with the specified labels added to any output.
//
// Use to create Logs specific to a particular lib or section of logic where
// the addtional labels can be used to identify that section in the logs
func (l *Log) WithLabels(labels ...any) *Log {
	sb := strings.Builder{}
	sb.Grow(1000)

	writeLabels(&sb, l.outputJSON, labels)

	return &Log{outputMask: l.outputMask, commonLabels: l.commonLabels + sb.String(), Writer: l.Writer}
}

// Writes a log with fatal severity and terminates the process
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	logger.Fatal(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func (l *Log) Fatal(ctx context.Context, message string, err error, labels ...any) {
	if l.outputMask&OutputFlagFatal == 0 {
		return
	}

	l.log(ctx, "FATAL", message, err, labels...)
	FatalFunc()
}

// Writes a log with error severity
// Error is reserved for events that causes a primary task to fail, such a http request or queued task handler
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	logger.Error(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func (l *Log) Error(ctx context.Context, message string, err error, labels ...any) {
	if l.outputMask&OutputFlagError == 0 {
		return
	}

	l.log(ctx, "ERROR", message, err, labels...)
}

// Writes a log with warning severity
// Warning is reserved for events that cause a process or task to run sub-optimally but not fail, such downstream API calls that fail
// initially but are retried and subsequently succeed
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	logger.Warning(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func (l *Log) Warning(ctx context.Context, message string, err error, labels ...any) {
	if l.outputMask&OutputFlagWarning == 0 {
		return
	}

	l.log(ctx, "WARNING", message, err, labels...)
}

// Writes a log with notice severity
// Notice is reserved for events that are expected but important, such as sytem start-up or shut-down
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	logger.Notice(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func (l *Log) Notice(ctx context.Context, message string, labels ...any) {
	if l.outputMask&OutputFlagNotice == 0 {
		return
	}

	l.log(ctx, "NOTICE", message, nil, labels...)
}

// Writes a log with info severity
// Info is reserved for emitting high level detail about a process's internal activity, such as completing a http request or queued task
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	logger.Info(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool)
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func (l *Log) Info(ctx context.Context, message string, labels ...any) {
	if l.outputMask&OutputFlagInfo == 0 {
		return
	}

	l.log(ctx, "INFO", message, nil, labels...)
}

// Writes a log with debug severity and a label of trace=true
// Trace is reserved for emitting data about IO within a process's internal activity, such as the content of requests received or generated
//
// Any number of labels can be provided but they must be given in key, value pairs
// where each key is a string. Values may be of any type or expressed as a func() T.
//
// For example:
//
//	logger.Trace(ctx, "some helpful information", "key1", "value1", "key2", 2, "key3", func() string { return "lazy_value3" } )
//
// Where the value is a `func() T`, the func will not be evaluated unless the output verbosity
// is such that the log will be written. Use this to prevent needless evaluation
// of expensive expressions (supports func T where T is string, int, uint, floats and bool).
//
// If the variadic labels argument cannot be be interpretted as balanced key, value pairs, then
// a `#missing#` value will be silently appended to balance them and provide some opportunity for discovery
func (l *Log) Trace(ctx context.Context, message string, labels ...any) {
	if l.outputMask&OutputFlagTrace == 0 {
		return
	}

	l.log(ctx, "DEBUG", message, nil, append(labels, "trace", true)...)
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
func (l *Log) Debug(ctx context.Context, message string, labels ...any) {
	if l.outputMask&OutputFlagDebug == 0 {
		return
	}

	l.log(ctx, "DEBUG", message, nil, labels...)
}

func (l *Log) log(ctx context.Context, severity, message string, err error, labels ...any) {
	b := make([]byte, 0, 500)

	openLog, closeLog, openField, closeField := `{ "`, ` }`, `, "`, `": `

	if !l.outputJSON {
		openLog, closeLog, openField, closeField = ``, ``, ` `, `=`
	}

	b = append(b, []byte(openLog+TraceIDFieldName+closeField+`"`+TraceID(ctx))...)
	b = append(b, []byte(`"`+openField+"severity"+closeField+`"`+severity)...)
	b = append(b, []byte(`"`+openField+"timestamp"+closeField+`"`)...)
	b = timeNow().UTC().AppendFormat(b, TimestampFormat)
	b = append(b, []byte(`"`)...)

	if err != nil {
		s := err.Error()

		if strings.Contains(err.Error(), `"`) {
			s = strings.ReplaceAll(s, `"`, `\"`)
		}

		b = append(b, []byte(`"`+openField+"error"+closeField+`"`+s+`"`)...)
	}

	b = append(b, []byte(l.commonLabels)...)

	// this is similar code to that in writeLabels(...) however it works on a []byte rather a strings.Builder
	// it is redefined inline to minimise the conditions when []byte must be allocated on the heap
	if len(labels)%2 != 0 {
		labels = append(labels, "#missing#")
	}

	for i := 0; i < len(labels); i += 2 {
		key, ok := labels[i].(string)

		if !ok {
			key = fmt.Sprintf("%v", labels[i])
		}

		val := ""
		switch v := labels[i+1].(type) {
		case string:
			val = `"` + v + `"`
		case int:
			val = strconv.Itoa(v)
		case uint:
			val = strconv.FormatUint(uint64(v), 10)
		case bool:
			val = strconv.FormatBool(v)
		case float32:
			val = strconv.FormatFloat(float64(v), 'f', 2, 64)
		case float64:
			val = strconv.FormatFloat(v, 'f', 2, 64)
		case fmt.Stringer:
			val = v.String()
		case func() string:
			val = `"` + v() + `"`
		case func() int:
			val = strconv.Itoa(v())
		case func() uint:
			val = strconv.FormatUint(uint64(v()), 10)
		case func() bool:
			val = strconv.FormatBool(v())
		case func() float32:
			val = strconv.FormatFloat(float64(v()), 'f', 2, 64)
		case func() float64:
			val = strconv.FormatFloat(v(), 'f', 2, 64)
		default: // handle the common primitives explicitly, accept an allocation or so for the rest and let fmt work its magic
			val = fmt.Sprintf("%v", labels[i+1])
		}

		b = append(b, []byte(openField+key+closeField+val)...)
	}

	b = append(b, []byte(openField+"message"+closeField+`"`)...)

	if strings.Contains(message, `"`) {
		message = strings.ReplaceAll(message, `"`, `\"`)
	}

	b = append(b, []byte(message)...)
	b = append(b, []byte(`"`+closeLog+"\n")...)

	mx.Lock()
	defer mx.Unlock()
	l.Writer.Write(b)
}

func writeLabels(sb *strings.Builder, outputJSON bool, labels []any) {
	if len(labels)%2 != 0 {
		labels = append(labels, "#missing#")
	}

	openField, closeField := `, "`, `": `

	if !outputJSON {
		openField, closeField = ` `, `=`
	}

	for i := 0; i < len(labels); i += 2 {
		key, ok := labels[i].(string)

		if !ok {
			continue
		}

		val := ""
		switch v := labels[i+1].(type) {
		case string:
			val = `"` + v + `"`
		case int:
			val = strconv.Itoa(v)
		case uint:
			val = strconv.FormatUint(uint64(v), 10)
		case bool:
			val = strconv.FormatBool(v)
		case float32:
			val = strconv.FormatFloat(float64(v), 'f', 2, 64)
		case float64:
			val = strconv.FormatFloat(v, 'f', 2, 64)
		case fmt.Stringer:
			val = v.String()
		default: // handle the common primitives explicitly, accept an allocation or so for the rest and let fmt work its magic
			val = fmt.Sprintf("%v", labels[i+1])
		}

		sb.WriteString(openField + key + closeField + val)
	}
}
