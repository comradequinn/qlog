package qlog

import (
	"context"
	"fmt"
	"io"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slog"
)

func BenchmarkQlog(b *testing.B) {
	SetOutputJSON(false)
	SetOutputMask(OutputFlagError)
	SetWriter(io.Discard)

	ctx := ContextFrom(context.Background(), "")
	err := fmt.Errorf("test error")

	b.Run("small-log", func(b *testing.B) {
		msg := "test message"
		labels := []any{"key1", "value1", "key2", "value2", "key3", func() string { return "lazyvalue3" }}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Error(ctx, msg, err, labels...)
		}
	})

	b.Run("medium-log", func(b *testing.B) {
		msg := "medium test message much larger than the small test message but not as large as the large message"
		labels := []any{"key1", "value1", "key2", 2, "key3", func() string { return "lazyvalue3" }, "key4", 3.14159, "key5", true, "key6", "value1", "key7", 2}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Error(ctx, msg, err, labels...)
		}
	})

	b.Run("large-log", func(b *testing.B) {
		msg := "large test message larger than the medium test message and much larger than the small test message. to repeat, this log is a large test message larger than the medium test message and much larger than the small test message"
		labels := []any{"key1", "value1", "key2", 2, "key3", func() string { return "lazyvalue3" }, "key4", 3.14159, "key5", true, "key6", "value1", "key7", 2, "key8", func() int { return 8 }, "key9", 3.14159, "key10", true}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Error(ctx, msg, err, labels...)
		}
	})

	b.Run("very-large-log", func(b *testing.B) {
		msg := "very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat again, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. and a final time, to be clear, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message"
		labels := []any{"key1", "value1", "key2", 2, "key3", func() string { return "lazyvalue3" }, "key4", 3.14159, "key5", true, "key6", "value1", "key7", 2, "key8", func() int { return 8 }, "key9", 3.14159, "key10", true}
		labels = append(labels, []any{"key10", "value1", "key12", 2, "key13", func() string { return "lazyvalue3" }, "key14", 3.14159, "key15", true, "key16", "value1", "key17", 2, "key18", func() int { return 8 }, "key19", 3.14159, "key20", true})
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Error(ctx, msg, err, labels...)
		}
	})
}

type LazyString func() string

func (l LazyString) LogValue() slog.Value {
	return slog.StringValue(l())
}

type LazyInt func() int

func (l LazyInt) LogValue() slog.Value {
	return slog.IntValue(l())
}

func BenchmarkSlog(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard))
	slog.SetDefault(logger)

	ctx := ContextFrom(context.Background(), "")
	err := fmt.Errorf("test error")

	b.Run("small-log", func(b *testing.B) {
		msg := "test message"
		labels := []any{"error", err.Error(), "key1", "value1", "key2", "value2", "key3", LazyString(func() string { return "lazyvalue3" })}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.ErrorCtx(ctx, msg, labels...)
		}
	})

	b.Run("medium-log", func(b *testing.B) {
		msg := "medium test message much larger than the small test message but not as large as the large message"
		labels := []any{"error", err.Error(), "key1", "value1", "key2", 2, "key3", LazyString(func() string { return "lazyvalue3" }), "key4", 3.14159, "key5", true, "key6", "value1", "key7", 2}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.ErrorCtx(ctx, msg, labels...)
		}
	})

	b.Run("large-log", func(b *testing.B) {
		msg := "large test message larger than the medium test message and much larger than the small test message. to repeat, this log is a large test message larger than the medium test message and much larger than the small test message"
		labels := []any{"error", err.Error(), "key1", "value1", "key2", 2, "key3", LazyString(func() string { return "lazyvalue3" }), "key4", 3.14159, "key5", true, "key6", "value1", "key7", 2, "key8", LazyInt(func() int { return 8 }), "key9", 3.14159, "key10", true}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.ErrorCtx(ctx, msg, labels...)
		}
	})

	b.Run("very-large-log", func(b *testing.B) {
		msg := "very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat again, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. and a final time, to be clear, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message"
		labels := []any{"error", err.Error(), "key1", "value1", "key2", 2, "key3", LazyString(func() string { return "lazyvalue3" }), "key4", 3.14159, "key5", true, "key6", "value1", "key7", 2, "key8", LazyInt(func() int { return 8 }), "key9", 3.14159, "key10", true}
		labels = append(labels, []any{"key11", "value1", "key12", 2, "key13", LazyString(func() string { return "lazyvalue3" }), "key14", 3.14159, "key15", true, "key16", "value1", "key17", 2, "key18", LazyInt(func() int { return 8 }), "key19", 3.14159, "key20", true})
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.ErrorCtx(ctx, msg, labels...)
		}
	})
}

func BenchmarkSlogLogAttr(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard))
	slog.SetDefault(logger)

	ctx := ContextFrom(context.Background(), "")
	err := fmt.Errorf("test error")

	b.Run("small-log", func(b *testing.B) {
		msg := "test message"
		labels := []slog.Attr{slog.String("error", err.Error()),
			slog.String("key1", "value1"),
			slog.String("key2", "value2"),
			slog.Any("key3", LazyString(func() string { return "lazyvalue3" }))}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.LogAttrs(ctx, slog.LevelError, msg, labels...)
		}
	})

	b.Run("medium-log", func(b *testing.B) {
		msg := "medium test message much larger than the small test message but not as large as the large message"
		labels := []slog.Attr{slog.String("error", err.Error()),
			slog.String("key1", "value1"),
			slog.String("key2", "value2"),
			slog.Any("key3", LazyString(func() string { return "lazyvalue3" })),
			slog.Float64("key4", 3.14159),
			slog.Bool("key6", true),
			slog.Int("key7", 2)}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.LogAttrs(ctx, slog.LevelError, msg, labels...)
		}
	})

	b.Run("large-log", func(b *testing.B) {
		msg := "large test message larger than the medium test message and much larger than the small test message. to repeat, this log is a large test message larger than the medium test message and much larger than the small test message"
		labels := []slog.Attr{slog.String("error", err.Error()),
			slog.String("key1", "value1"),
			slog.String("key2", "value2"),
			slog.Any("key3", LazyString(func() string { return "lazyvalue3" })),
			slog.Float64("key4", 3.14159),
			slog.Bool("key6", true),
			slog.Int("key7", 2),
			slog.Any("key8", LazyInt(func() int { return 8 })),
			slog.Float64("key9", 3.14159),
			slog.Bool("key10", true)}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.LogAttrs(ctx, slog.LevelError, msg, labels...)
		}
	})

	b.Run("very-large-log", func(b *testing.B) {
		msg := "very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat again, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. and a final time, to be clear, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message"
		labels := []slog.Attr{slog.String("error", err.Error()),
			slog.String("key1", "value1"),
			slog.String("key2", "value2"),
			slog.Any("key3", LazyString(func() string { return "lazyvalue3" })),
			slog.Float64("key4", 3.14159),
			slog.Bool("key6", true),
			slog.Int("key7", 2),
			slog.Any("key8", LazyInt(func() int { return 8 })),
			slog.Float64("key9", 3.14159),
			slog.Bool("key10", true),
			slog.String("key11", "value1"),
			slog.String("key12", "value2"),
			slog.Any("key13", LazyString(func() string { return "lazyvalue3" })),
			slog.Float64("key14", 3.14159),
			slog.Bool("key15", true),
			slog.String("key16", "value1"),
			slog.Int("key17", 2),
			slog.Any("key18", LazyInt(func() int { return 8 })),
			slog.Float64("key19", 3.14159),
			slog.Bool("key20", true)}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.LogAttrs(ctx, slog.LevelError, msg, labels...)
		}
	})
}

func BenchmarkZapSugared(b *testing.B) {
	zlog := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.ErrorLevel))
	slog := zlog.Sugar()
	err := fmt.Errorf("test error")

	b.Run("small-log", func(b *testing.B) {
		msg := "test message"
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.Error(
				"msg", msg,
				"error", err.Error(),
				"key1", "value1",
				"key2", "value2",
				"key3", func() any { return "lazyvalue3" }(),
			)
		}

		zlog.Sync()
	})

	b.Run("medium-log", func(b *testing.B) {
		msg := "medium test message much larger than the small test message but not as large as the large message"
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.Error(
				"msg", msg,
				"error", err.Error(),
				"key1", "value1",
				"key2", "value2",
				"key3", func() any { return "lazyvalue3" }(), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
				"key4", 3.14159,
				"key6", true,
				"key7", 2,
			)
		}

		zlog.Sync()
	})

	b.Run("large-log", func(b *testing.B) {
		msg := "large test message larger than the medium test message and much larger than the small test message. to repeat, this log is a large test message larger than the medium test message and much larger than the small test message"
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			slog.Error(
				"msg", msg,
				"error", err.Error(),
				"key1", "value1",
				"key2", "value2",
				"key3", func() any { return "lazyvalue3" }(), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
				"key4", 3.14159,
				"key6", true,
				"key7", 2,
				"key8", func() any { return 8 }(),
				"key9", 3.14159,
				"key10", true,
			)
		}

		zlog.Sync()
	})

	b.Run("very-large-log", func(b *testing.B) {
		msg := "very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat again, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. and a final time, to be clear, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message"
		for i := 0; i < b.N; i++ {
			slog.Error(
				"msg", msg,
				"error", err.Error(),
				"key1", "value1",
				"key2", "value2",
				"key3", func() any { return "lazyvalue3" }(), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
				"key4", 3.14159,
				"key6", true,
				"key7", 2,
				"key8", func() any { return 8 }(),
				"key9", 3.14159,
				"key10", true,
				"key11", "value1",
				"key12", "value2",
				"key13", func() any { return "lazyvalue3" }(),
				"key14", 3.14159,
				"key15", true,
				"key16", "value1",
				"key17", 2,
				"key18", func() any { return 8 }(),
				"key19", 3.14159,
				"key20", true,
			)
		}

		slog.Sync()
	})
}

func BenchmarkZapUnsugared(b *testing.B) {
	zlog := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(io.Discard), zapcore.ErrorLevel))
	err := fmt.Errorf("test error")

	b.Run("small-log", func(b *testing.B) {
		msg := "test message"
		labels := []zap.Field{zap.String("error", err.Error()),
			zap.String("key1", "value1"),
			zap.String("key2", "value2"),
			zap.Any("key3", func() any { return "lazyvalue3" }()), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			zlog.Error(msg, labels...)
		}

		zlog.Sync()
	})

	b.Run("medium-log", func(b *testing.B) {
		msg := "medium test message much larger than the small test message but not as large as the large message"
		labels := []zap.Field{zap.String("error", err.Error()),
			zap.String("key1", "value1"),
			zap.String("key2", "value2"),
			zap.Any("key3", func() any { return "lazyvalue3" }()), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
			zap.Float64("key4", 3.14159),
			zap.Bool("key6", true),
			zap.Int("key7", 2),
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			zlog.Error(msg, labels...)
		}

		zlog.Sync()
	})

	b.Run("large-log", func(b *testing.B) {
		msg := "large test message larger than the medium test message and much larger than the small test message. to repeat, this log is a large test message larger than the medium test message and much larger than the small test message"
		labels := []zap.Field{zap.String("error", err.Error()),
			zap.String("key1", "value1"),
			zap.String("key2", "value2"),
			zap.Any("key3", func() any { return "lazyvalue3" }()), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
			zap.Float64("key4", 3.14159),
			zap.Bool("key6", true),
			zap.Int("key7", 2),
			zap.Any("key8", func() any { return 8 }),
			zap.Float64("key9", 3.14159),
			zap.Bool("key10", true),
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			zlog.Error(msg, labels...)
		}

		zlog.Sync()
	})

	b.Run("very-large-log", func(b *testing.B) {
		msg := "very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. to repeat again, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message. and a final time, to be clear, very large test message. larger than the large test message and larger than the medium test message and much larger than the small test message"
		labels := []zap.Field{zap.String("error", err.Error()),
			zap.String("key1", "value1"),
			zap.String("key2", "value2"),
			zap.Any("key3", func() any { return "lazyvalue3" }()), // NB: Cannot find an implicit way of deferring execution in zap, this seems a reasonable equivalent in resource usage
			zap.Float64("key4", 3.14159),
			zap.Bool("key6", true),
			zap.Int("key7", 2),
			zap.Any("key8", func() any { return 8 }),
			zap.Float64("key9", 3.14159),
			zap.Bool("key10", true),
			zap.String("key11", "value1"),
			zap.String("key12", "value2"),
			zap.Any("key13", func() any { return "lazyvalue3" }()),
			zap.Float64("key14", 3.14159),
			zap.Bool("key15", true),
			zap.String("key16", "value1"),
			zap.Int("key17", 2),
			zap.Any("key18", func() any { return 8 }()),
			zap.Float64("key19", 3.14159),
			zap.Bool("key20", true)}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			zlog.Error(msg, labels...)
		}

		zlog.Sync()
	})
}
