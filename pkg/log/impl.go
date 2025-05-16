package log

import (
	"context"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/lengzuo/fundflow/configs"
	"github.com/rs/zerolog"
)

type ctxKey int

const (
	UsernameKey ctxKey = iota
)

type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Fatal(ctx context.Context, msg string, args ...any)
}

type Field map[string]any

type zerologWrapper struct {
	z zerolog.Logger
}

var logInstance Logger = &noLogger{}

func New(mode configs.Mode) {
	level := zerolog.DebugLevel
	if mode == configs.Prod {
		level = zerolog.InfoLevel
	}
	zerolog.TimeFieldFormat = time.RFC3339
	z := zerolog.New(os.Stdout).With().CallerWithSkipFrameCount(4).Timestamp().Caller().Logger().Level(level)
	logInstance = &zerologWrapper{z: z}
}

func (l *zerologWrapper) Debug(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, l.z.Debug).Msgf(msg, args...)
}
func (l *zerologWrapper) Info(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, l.z.Info).Msgf(msg, args...)
}
func (l *zerologWrapper) Warn(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, l.z.Warn).Msgf(msg, args...)
}
func (l *zerologWrapper) Error(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, l.z.Error).Msgf(msg, args...)
}
func (l *zerologWrapper) Fatal(ctx context.Context, msg string, args ...any) {
	l.logWithContext(ctx, l.z.Fatal).Msgf(msg, args...)
}

func (l *zerologWrapper) logWithContext(ctx context.Context, level func() *zerolog.Event) *zerolog.Event {
	e := level()
	if reqID, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		e.Str("request_id", reqID)
	}
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		e.Str("username", username)
	}
	return e
}
