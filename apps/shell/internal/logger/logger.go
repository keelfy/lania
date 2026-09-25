package logger

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/lania-smp/shell/internal/config"
)

const levelFatal = slog.Level(12)

// jsonLogger is set outside debug mode so production logs are structured for Loki.
var jsonLogger *slog.Logger

func PrepareLogger() {
	if !config.IsDebug() {
		jsonLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey && a.Value.Any() == levelFatal {
					a.Value = slog.StringValue("FATAL")
				}
				return a
			},
		}))
		return
	}

	log.SetFlags(log.LstdFlags)
	color.NoColor = false // enable color output
}

func printJSON(ctx context.Context, level, message string) {
	var slogLevel slog.Level
	switch level {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "WARN":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	case "FATAL":
		slogLevel = levelFatal
	default:
		slogLevel = slog.LevelInfo
	}

	jsonLogger.Log(ctx, slogLevel, message)

	if level == "FATAL" {
		os.Exit(1)
	}
}

func println(ctx context.Context, level, message string) {
	if jsonLogger != nil {
		printJSON(ctx, level, message)
		return
	}

	var colorFunc func(format string, a ...any) string

	switch level {
	case "INFO":
		colorFunc = color.New(color.FgGreen).SprintfFunc()
	case "DEBUG":
		colorFunc = color.New(color.FgBlue).SprintfFunc()
	case "ERROR":
		colorFunc = color.New(color.FgRed).SprintfFunc()
	case "WARN":
		colorFunc = color.New(color.FgYellow).SprintfFunc()
	case "FATAL":
		colorFunc = color.New(color.FgHiRed).SprintfFunc()
	default:
		colorFunc = color.New(color.FgWhite).SprintfFunc()
	}

	log.Println(fmt.Sprintf("%s %s", colorFunc("[%s]", level), message))

	if level == "FATAL" {
		log.Fatalln("Exiting...")
	}
}

func printf(ctx context.Context, level, message string, v ...any) {
	println(ctx, level, fmt.Sprintf(message, v...))
}

func Infof(ctx context.Context, message string, v ...any) {
	printf(ctx, "INFO", message, v...)
}

func Debugf(ctx context.Context, message string, v ...any) {
	if config.IsDebug() {
		printf(ctx, "DEBUG", message, v...)
	}
}

func Warnf(ctx context.Context, message string, v ...any) {
	printf(ctx, "WARN", message, v...)
}

func Errorf(ctx context.Context, message string, v ...any) {
	printf(ctx, "ERROR", message, v...)
}

func Fatalf(ctx context.Context, message string, v ...any) {
	printf(ctx, "FATAL", message, v...)
}
