package zap

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

func tryAddRequestID(req *http.Request) (reqid string) {
	reqid = GetRequestID(req.Context())
	if reqid != "" {
		return
	}
	u1, _ := uuid.NewV7()
	reqid = u1.String()
	*req = *req.WithContext(context.WithValue(req.Context(), ctxKeyRequestID{}, reqid))
	return
}

func GetRequestID(ctx context.Context) string {
	val, ok := ctx.Value(ctxKeyRequestID{}).(string)
	if !ok {
		return ""
	}
	return val
}
