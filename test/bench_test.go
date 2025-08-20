package zeroslog_test

import (
	"errors"
	"github.com/calyrexx/zeroslog"
	"io"
	"log/slog"
	"testing"
	"time"
)

var (
	fullMethod = "/test.TestService/GetUser"
	dur        = 320 * time.Microsecond
	version    = 21
	someFloat  = 123.123
	errTest    = errors.New("this is an error")
	timeFormat = "2006-01-02 15:04:05.000 -07:00"
)

func BenchmarkSlog_Info(b *testing.B) {
	slogLogger := slog.New(
		slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	benchLog(b, func() {
		slogLogger.Info("gRPC call succeeded",
			"method", fullMethod,
			"duration", dur,
			"version", version,
			"someFloat", someFloat,
		)
	})
}

func BenchmarkZeroSLog_Info(b *testing.B) {
	logger := slog.New(zeroslog.New(
		zeroslog.WithTimeFormat(timeFormat),
		zeroslog.WithOutput(io.Discard),
		zeroslog.WithColors(),
		zeroslog.WithMinLevel(slog.LevelInfo),
	))

	benchLog(b, func() {
		logger.Info("gRPC call succeeded",
			"method", fullMethod,
			"duration", dur,
			"version", version,
			"someFloat", someFloat,
		)
	})
}

func BenchmarkSlog_Error(b *testing.B) {
	slogLogger := slog.New(
		slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	benchLog(b, func() {
		slogLogger.Error("gRPC call succeeded",
			"method", fullMethod,
			"duration", dur,
			"err", errTest,
		)
	})
}

func BenchmarkZeroSLog_Error(b *testing.B) {
	logger := slog.New(zeroslog.New(
		zeroslog.WithTimeFormat(timeFormat),
		zeroslog.WithOutput(io.Discard),
		zeroslog.WithColors(),
		zeroslog.WithMinLevel(slog.LevelInfo),
	))

	benchLog(b, func() {
		logger.Error("gRPC call succeeded",
			"method", fullMethod,
			"duration", dur,
			"err", errTest,
		)
	})
}

func BenchmarkSlog_Groups(b *testing.B) {
	slogLogger := slog.New(
		slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	benchLog(b, func() {
		slogLogger.Info("gRPC call succeeded",
			"method", fullMethod,
			"duration", dur,
			slog.Group("user", "name", "John Doe", "id", 123456789),
		)
	})
}

func BenchmarkZeroSLog_Groups(b *testing.B) {
	logger := slog.New(zeroslog.New(
		zeroslog.WithTimeFormat(timeFormat),
		zeroslog.WithOutput(io.Discard),
		zeroslog.WithColors(),
		zeroslog.WithMinLevel(slog.LevelInfo),
	))

	benchLog(b, func() {
		logger.Info("gRPC call succeeded",
			"method", fullMethod,
			"duration", dur,
			slog.Group("user", "name", "John Doe", "id", 123456789),
		)
	})
}

func benchLog(b *testing.B, fn func()) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}
