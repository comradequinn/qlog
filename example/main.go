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
	text := flag.Bool("logfmt", false, "whether to output the log as logfmt instead of json")

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
		w.Header().Set("Span-ID", qlog.TraceID(ctx))

		// Write an informational log.
		// Note that as URL is passed as a `func() string` not a `string` it is  only resolved if the log is actually written, ie, if info level logging is enabled.
		// Use this to avoid costly expression evaluations that may not be needed if lower severity logging is not enabled (can be used with string, int, uint, floats and bool)
		qlog.Info(ctx, "received echo request", "url", func() string { return r.URL.String() }, "origin", r.RemoteAddr)

		if _, err := fmt.Fprintf(w, "echo: %v\n", r.URL.Query().Get("data")); err != nil {
			qlog.Error(ctx, "error processing request", err)
		}
	})

	qlog.Notice(ctx, `http "server" listening`) // record a notice in the log regarding the process starting

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		qlog.Fatal(ctx, "unable to start http server", err) // log the error and terminate the process
	}

	select {}
}
