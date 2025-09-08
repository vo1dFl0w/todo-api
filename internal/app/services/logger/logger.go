package logger

import (
	"log/slog"
	"os"
	"time"
)

var (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func InitLogger(env string) *slog.Logger {
	var log *slog.Logger

	options := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if v := a.Value.Any(); v != nil {
				if a.Key == slog.TimeKey {
					if t, ok := v.(time.Time); ok {
						return slog.String(slog.TimeKey, t.Format("02/01/2006 15:04:05"))
					}
				}
			}
			return a
		},
	}

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: options.ReplaceAttr}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: options.ReplaceAttr}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: options.ReplaceAttr}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, ReplaceAttr: options.ReplaceAttr}))
	}

	return log
}
