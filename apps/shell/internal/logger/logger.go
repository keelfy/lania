package logger

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/lania-smp/shell/internal/config"
)

func PrepareLogger() {
	log.SetFlags(log.LstdFlags)
	color.NoColor = false // enable color output
}

func println(_ context.Context, level, message string) {
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
