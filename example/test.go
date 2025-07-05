package main

import (
	"github.com/calyrexx/zeroslog"
	"log/slog"
	"time"
)

func main() {
	handler := zeroslog.New(
		zeroslog.WithMinLevel(slog.LevelDebug),
	)

	logger := slog.New(handler)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
		logger.Info("Сервер запущен",
			"port", 8080,
			"env", "production",
		)
	}
}
