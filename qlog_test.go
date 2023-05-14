package qlog

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestOutput(t *testing.T) {
	testError := fmt.Errorf("test error")
	tcs := []struct {
		Desc        string
		Severity    string
		OutputMask  int
		TargetFunc  func(context.Context, string, ...any)
		ExpectEmpty bool
		ExtraLabels []any
	}{
		{
			Desc:        "TestFatal",
			Severity:    "FATAL",
			OutputMask:  OutputFlagFatal,
			TargetFunc:  func(ctx context.Context, s string, a ...any) { Fatal(ctx, s, testError, a...) },
			ExtraLabels: []any{"error", `"` + testError.Error() + `"`},
		},
		{
			Desc:        "TestFatalDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  func(ctx context.Context, s string, a ...any) { Fatal(ctx, s, testError, a...) },
			ExpectEmpty: true,
		},
		{
			Desc:        "TestError",
			Severity:    "ERROR",
			OutputMask:  OutputFlagError,
			TargetFunc:  func(ctx context.Context, s string, a ...any) { Error(ctx, s, testError, a...) },
			ExtraLabels: []any{"error", `"` + testError.Error() + `"`},
		},
		{
			Desc:        "TestErrorDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  func(ctx context.Context, s string, a ...any) { Error(ctx, s, testError, a...) },
			ExpectEmpty: true,
		},
		{
			Desc:        "TestWarning",
			Severity:    "WARNING",
			OutputMask:  OutputFlagWarning,
			TargetFunc:  func(ctx context.Context, s string, a ...any) { Warning(ctx, s, testError, a...) },
			ExtraLabels: []any{"error", `"` + testError.Error() + `"`},
		},
		{
			Desc:        "TestWarningDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  func(ctx context.Context, s string, a ...any) { Warning(ctx, s, testError, a...) },
			ExpectEmpty: true,
		},
		{
			Desc:       "TestNotice",
			Severity:   "NOTICE",
			OutputMask: OutputFlagNotice,
			TargetFunc: Notice,
		},
		{
			Desc:        "TestNoticeDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  Notice,
			ExpectEmpty: true,
		},
		{
			Desc:        "TestTrace",
			Severity:    "DEBUG",
			OutputMask:  OutputFlagTrace,
			TargetFunc:  Trace,
			ExtraLabels: []any{"trace", "true"},
		},
		{
			Desc:        "TestTraceDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  Trace,
			ExpectEmpty: true,
		},
		{
			Desc:       "TestInfo",
			Severity:   "INFO",
			OutputMask: OutputFlagInfo,
			TargetFunc: Info,
		},
		{
			Desc:        "TestInfoDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  Info,
			ExpectEmpty: true,
		},
		{
			Desc:       "TestDebug",
			Severity:   "DEBUG",
			OutputMask: OutputFlagDebug,
			TargetFunc: Debug,
		},
		{
			Desc:        "TestDebugDisabled",
			OutputMask:  OutputFlagNone,
			TargetFunc:  Debug,
			ExpectEmpty: true,
		},
	}

	FatalFunc = func() {}
	expectedTime := time.Now()
	timeNow = func() time.Time { return expectedTime }

	assert := func(formatField func(k, v any) string) {
		SetLabels("common", true)

		for _, tc := range tcs {
			sb := strings.Builder{}

			SetWriter(&sb)

			SetOutputMask(tc.OutputMask)

			ctx := ContextFrom(context.Background(), "")
			msg := "test message"
			labels := []any{
				"stringkey", "stringval",
				"intkey", 2,
				"uintkey", uint(3),
				"boolkey", true,
				"float32key", float32(41.3),
				"float64key", 3.14,
				"stringfunckey", func() string { return "stringfuncval" },
				"intfunckey", func() int { return 4 },
				"uintfunckey", func() uint { return uint(6) },
				"boolfunckey", func() bool { return false },
				"float32funckey", func() float32 { return 81.3 },
				"float64funckey", func() float64 { return 6.14 },
			}

			tc.TargetFunc(ctx, msg, labels...)

			output := sb.String()

			if tc.ExpectEmpty {
				if output != "" {
					t.Fatalf("%v: expected empty output but got '%s'", tc.Desc, output)
				}
				continue
			}

			labels[1] = `"stringval"`
			labels[3] = `2`
			labels[5] = `3`
			labels[7] = `true`
			labels[9] = `41.3`
			labels[11] = `3.14`
			labels[13] = `"stringfuncval"`
			labels[15] = `4`
			labels[17] = `6`
			labels[19] = `false`
			labels[21] = `81.3`
			labels[23] = `6.14`

			expectedLabels := append(labels, "common", "true", "severity", `"`+tc.Severity+`"`, "trace", `"`+TraceID(ctx)+`"`, "message", `"`+msg+`"`, "timestamp", `"`+expectedTime.UTC().Format(time.RFC3339Nano)+`"`)
			expectedLabels = append(expectedLabels, tc.ExtraLabels...)

			for i := 0; i < len(expectedLabels); i += 2 {
				label := formatField(expectedLabels[i], expectedLabels[i+1])

				if !strings.Contains(strings.ToLower(output), strings.ToLower(label)) {
					t.Fatalf("%v: expected label '%s', but got '%s'", tc.Desc, label, output)
				}
			}
		}
	}

	assert(func(k, v any) string { return fmt.Sprintf(`"%s": %s`, k, v) })
	SetOutputJSON(false)
	assert(func(k, v any) string { return fmt.Sprintf(`%s=%s`, k, v) })
}
