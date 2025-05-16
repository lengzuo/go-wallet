package log

import (
	"context"
)

func Debug(ctx context.Context, msg string, args ...any) {
	logInstance.Debug(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	logInstance.Info(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	logInstance.Warn(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	logInstance.Error(ctx, msg, args...)
}

func Fatal(ctx context.Context, msg string, args ...any) {
	logInstance.Fatal(ctx, msg, args...)
}
