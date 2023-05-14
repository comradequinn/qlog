# qlog

A simple, fast, structured logging library with built-in task/request tracing.

## Features
* `Speed & GC Efficiency` : Compared to [slog](https://pkg.go.dev/golang.org/x/exp/slog), `qlog` performs considerably faster and almost* always requires fewer allocations  
* `Process Tracing` : Logs written with the same context are assigned the same, unique `Span ID`. This allows related logs to be easily collated  
* `Structured Labels` : Log labels are provided via a variadic parameter on each of the various log methods or specified once as common to all logs  
* `Fluid Interface` : Many loggers require labels to be wrapped in custom types to help reduce allocations, `qlog` takes any plain old go types
* `Deferred Evaluation` : By defining label values as funcs, evaluation will deferred until the log is written, avoiding costly evaluations for logs whose output is disabled
* `Format Control` : Logs can be written either as `JSON` or as logfmt
* `Verbosity Control` : Log output verbosity is controlled by configuring the OutputMask; either with the individual OutputFlags
required, or by using one of the preset OutputMasks

## Quick Start
To write a log, decide on the severity of the event to be recorded and call the appropriate func. 

To enable request/task tracing, create, or extend an existing, `Context` using the `qlog.ContextFrom(...)` func. The returned `Context` may then be used in related `log.*` calls to allow them to be grouped together in log analysis tools. 

```go
ctx := qlog.ContextFrom(context.Background(), "") // generate a new context that can be used to link related logs

qlog.Info(ctx, "processing request") // this log will have the same `Trace-ID` as...  
qlog.Error(ctx, "error processing request", err) // .. this log

ctx2 =: qlog.ContextFrom(context.Background(), "") 
qlog.Info(ctx2, "processing request") // this log will have a different Trace-ID as it was passed a different ctx
```

In addition to messages and errors, an arbitary numbers of labels can be added to logs expressed as key value pairs and passed as a variadic argument to the log method. The keys for these labels should be strings but the value may be of any type.

```go
qlog.Info(ctx, "received request", "port", 80, "origin", r.RemoteAddr)
```

In the event that an attribute value is expensive to evaluate, this may be deferred until the log is actually written (_meaning the evaluation does not occur if the log's severity is not included in the enabled ouput_). To do this, instead of passing the value `T` directly, define a `func() T` that evaluates and returns `T` when executed.

```go
// T may be any of `string`, `int`, `uint`, `floats` and `bool`
qlog.Info(ctx, "received request", "url", func() string { return r.URL.String() }, "port", 80)
```

Typically, a set of standard labels need including on every log. Rather than defining these on each `log.*` call, they can be set once and applied to all future logs

```go
qlog.SetLabels("app", "example", "port", *port)
```
 
 Depending on the environment that the system is executing in, different outputs may be required. `qlog` can be configured to output `JSON` or `logfmt` and each severity can be specifically included or excluded by using varying combinations of the provided `Output Masks` and `Output Flags`

 ```go
qlog.SetOutputJSON(false) // set the output to logfmt (json is the default)

qlog.SetOutputMask(qlog.OutputMaskAll) // use a pre-configured mask to output all logs
qlog.SetOutputMask(qlog.OutputMaskImportant) // use a pre-configured mask to output Fatal, Error, Warn and Notice logs
qlog.SetOutputMask(qlog.OutputFlagFatal|qlog.OutputFlagTrace) // use a custom mask that includes only Fatal and Trace logs
 ```

In some cases, rather than using a top level `qlog.*` func, a specific instance may be required with its own configuration. This is usually to tailor the logging to a particular subset of logic, perhaps by adding further labels, or to satisfy an interface. In either case, such instances may be created as shown below.

```go
log1 := qlog.New(qlog.OutputMaskAll, true, "section", "sensitive") // creates a new loger with a custom configuration
log2 := log1.WithLabels("subsection", "critical") // creates a new logger based on the current one, with added labels
```

## Why not use slog?

[slog](https://pkg.go.dev/golang.org/x/exp/slog) is an excellent logger, but for use-cases commonly encountered in many systems, `qlog` is simpler and more efficient. 

When compared to [slog](https://pkg.go.dev/golang.org/x/exp/slog), `qlog` is considerably faster under typical logging conditions. It almost* always requires fewer allocations than [slog](https://pkg.go.dev/golang.org/x/exp/slog) per log written. It has a far simpler interface, particularly when deferring evaluation of attribute values. 

In addition to this, request/task tracing which is commonly required in production scenarios, is built-in to `qlog` but requires further extension with [slog](https://pkg.go.dev/golang.org/x/exp/slog).

It should be stated, however, that [slog](https://pkg.go.dev/golang.org/x/exp/slog) does more in terms of the output formats it can generate and its general configurability, which accounts for its increased resource usage. 

Benchmarks for common logging scenarios comparing both [slog](https://pkg.go.dev/golang.org/x/exp/slog)'s funcs that are API-equivalent to `qlog` and it's performance func, `LogAttrs` are shown below (results are reflective of performance on amd64/linux and arm64/mac). 

```bash
# qlog tests benchmarks
BenchmarkQlog/small-log-8                2388529               493.6 ns/op           512 B/op          1 allocs/op
BenchmarkQlog/medium-log-8               1549069               768.7 ns/op           652 B/op          4 allocs/op
BenchmarkQlog/large-log-8                1000000              1109 ns/op            1688 B/op          7 allocs/op
BenchmarkQlog/very-large-log-8            461336              2529 ns/op            2825 B/op         10 allocs/op

# slog's API-equivalent methods benchmarks
BenchmarkSlog/small-log-8                1615983               739.5 ns/op             0 B/op          0 allocs/op
BenchmarkSlog/medium-log-8                826692              1292 ns/op             360 B/op          5 allocs/op
BenchmarkSlog/large-log-8                 662647              1777 ns/op             864 B/op          8 allocs/op
BenchmarkSlog/very-large-log-8            357340              3305 ns/op            1137 B/op         13 allocs/op

# slog's performance method (LogAttrs) benchmarks
BenchmarkSlogLogAttr/small-log-8         1540220               775.5 ns/op             0 B/op          0 allocs/op
BenchmarkSlogLogAttr/medium-log-8         954751              1261 ns/op             200 B/op          4 allocs/op
BenchmarkSlogLogAttr/large-log-8          662775              1729 ns/op             448 B/op          7 allocs/op
BenchmarkSlogLogAttr/very-large-log-8     381471              3060 ns/op            1120 B/op         13 allocs/op
```

*_Note that the log size at which [slog](https://pkg.go.dev/golang.org/x/exp/slog)'s funcs do not allocate at all is quite small. It consists of a two word message and just three labels. At this size `qlog` allocates once but is faster. However, with more realistic log sizes (a single sentence and 5+ labels), both [slog](https://pkg.go.dev/golang.org/x/exp/slog) and `qlog` now allocate: but, `qlog` allocates fewer times, and yet still remains considerably faster. This trend continues as log size increases._

## Example
An example application is shown below and available in the repo. It can be executed with `make example`

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/comradequinn/qlog"
)

func main() {
	port := flag.Int("port", 8080, "the port to listen on")
	text := flag.Bool("logfmt", false, "whether to output the log as plogfmt instead of json")

	flag.Parse()

	qlog.SetOutputJSON(!(*text))
	qlog.SetOutputMask(qlog.OutputMaskAll) // use a pre-configured mask to output all logs

	// Add common labels that will be included in all logs, any non-func type can be specified
	qlog.SetLabels("app", "example", "port", *port)

	ctx := qlog.ContextFrom(context.Background(), "")

	http.HandleFunc("/echo/", func(w http.ResponseWriter, r *http.Request) {
		// Create a custom context for this request, all logs generated with this ctx will have the same Trace-ID.
		// If the header contains a Trace-ID then the client and server logs can be linked across service boundaries.
		// If header is missing, the empty string passed will cause a new Trace-ID to be generated
		ctx := qlog.ContextFrom(ctx, r.Header.Get("Span-ID"))

		// Add the Trace-ID to the response headers so that clients may link their own logs
		w.Header().Set("Span-ID", qlog.SpanID(ctx))

		// Write an informational log.
		// Note that as URL is passed as a `func() string` not a `string` it is  only resolved if the log is actually written, ie, if info level logging is enabled.
		// Use this to avoid costly expression evaluations that may not be needed if lower severity logging is not enabled (can be used with string, int, uint, floats and bool)
		qlog.Info(ctx, "received echo request", "url", func() string { return r.URL.String() }, "origin", r.RemoteAddr)

		if _, err := fmt.Fprintf(w, "echo: %v\n", r.URL.Query().Get("data")); err != nil {
			qlog.Error(ctx, "error processing request", err)
		}
	})

	qlog.Notice(ctx, "http server listening") // record a notice in the log regarding the process starting

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		qlog.Fatal(ctx, "unable to start http server", err) // log the error and terminate the process
	}

	select {}
}
```

Sample output from the above is shown below:

```json
{ "trace": "1349668033305042412", "severity": "NOTICE", "timestamp": "2023-05-13T16:37:48.214451Z", "app": "example", "port": 8080, "port": 8080, "message": "http server listening" }
{ "trace": "1691594436394702499", "severity": "INFO", "timestamp": "2023-05-13T16:37:48.214452Z", "app": "example", "port": 8080, "url": "/echo/?data=hello", "origin": "127.0.0.1:54944", "message": "received echo request" }
{ "trace": "2270764459917579688", "severity": "INFO", "timestamp": "2023-05-13T16:37:48.214453Z", "app": "example", "port": 8080, "url": "/echo/?data=world", "origin": "127.0.0.1:54948", "message": "received echo request" }
```

