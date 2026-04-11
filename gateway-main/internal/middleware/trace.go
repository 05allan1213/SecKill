package middleware

import (
	"fmt"
	"net/http"
	"time"

	gwlog "github.com/BitofferHub/gateway/internal/log"
	"github.com/BitofferHub/pkg/constant"
)

type TraceMiddleware struct{}

func NewTraceMiddleware() *TraceMiddleware {
	return &TraceMiddleware{}
}

func (m *TraceMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(constant.TraceID)
		if traceID == "" {
			traceID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		w.Header().Set(constant.TraceID, traceID)
		ctx := WithTraceID(r.Context(), traceID)
		ctx = gwlog.WithTrace(ctx, traceID)
		next(w, r.WithContext(ctx))
	}
}
