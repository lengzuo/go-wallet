package log

import "context"

type noLogger struct{}

func (n *noLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (n *noLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (n *noLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (n *noLogger) Error(ctx context.Context, msg string, args ...any) {}
func (n *noLogger) Fatal(ctx context.Context, msg string, args ...any) {}
