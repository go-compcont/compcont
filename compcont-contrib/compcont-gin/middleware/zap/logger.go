package zap

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

var (
	defaultLogger **zap.Logger
)

type ctxKeyLogger struct{}

func WithLogger(req *http.Request, logger *zap.Logger) {
	ctx := req.Context()
	if _, ok := ctx.Value(ctxKeyLogger{}).(*zap.Logger); ok {
		return
	}
	*req = *req.WithContext(context.WithValue(ctx, ctxKeyLogger{}, logger))
}

func FromContext(ctx context.Context) *zap.Logger {
	val, ok := ctx.Value(ctxKeyLogger{}).(*zap.Logger)
	if !ok {
		return GetDefault()
	}
	return val
}

func SetDefault(l *zap.Logger) {
	defaultLogger = &l
}

func GetDefault() *zap.Logger {
	if defaultLogger == nil {
		return zap.NewNop()
	} else {
		return *defaultLogger
	}
}
