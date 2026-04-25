package logger

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/lania-smp/backend/internal/config"
)

func PrepareLogger() {
	log.SetFlags(log.LstdFlags)
	color.NoColor = false // enable color output
}

func println(ctx context.Context, level, message string) {
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

	coloredLevel := colorFunc("[%s]", level)

	requestId := chiMiddleware.GetReqID(ctx)
	if requestId == "" {
		requestId = "—"
	}

	coloredRequestId := color.New(color.FgYellow).Sprintf("[%s]", requestId)

	log.Println(fmt.Sprintf("%s %s %s", coloredRequestId, coloredLevel, message))

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

func Info(ctx context.Context, message string) {
	println(ctx, "INFO", message)
}

func Debugf(ctx context.Context, message string, v ...any) {
	if config.IsDebug() {
		printf(ctx, "DEBUG", message, v...)
	}
}

func Debug(ctx context.Context, message string) {
	if config.IsDebug() {
		println(ctx, "DEBUG", message)
	}
}

func Errorf(ctx context.Context, message string, v ...any) {
	printf(ctx, "ERROR", message, v...)
}

func Error(ctx context.Context, message string) {
	println(ctx, "ERROR", message)
}

func Warnf(ctx context.Context, message string, v ...any) {
	printf(ctx, "WARN", message, v...)
}

func Warn(ctx context.Context, message string) {
	println(ctx, "WARN", message)
}

func Fatalf(ctx context.Context, message string, v ...any) {
	printf(ctx, "FATAL", message, v...)
}

func Fatal(ctx context.Context, message string) {
	println(ctx, "FATAL", message)
}
