package middleware

import (
	"net/http"

	gwlog "github.com/BitofferHub/gateway/internal/log"
	"github.com/BitofferHub/gateway/internal/svc"
)

const rateLimitCode = 42901

type RouteLimitMiddleware struct {
	svcCtx   *svc.ServiceContext
	routeKey string
}

func NewRouteLimitMiddleware(svcCtx *svc.ServiceContext, routeKey string) *RouteLimitMiddleware {
	return &RouteLimitMiddleware{
		svcCtx:   svcCtx,
		routeKey: routeKey,
	}
}

func (m *RouteLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.svcCtx.Limiter == nil {
			next(w, r)
			return
		}

		result, err := m.svcCtx.Limiter.Allow(r.Context(), m.routeKey)
		if err != nil || !result.IsAllowed {
			payload := map[string]interface{}{
				"code":    rateLimitCode,
				"message": "rate limited",
			}
			RecordAccessCode(r.Context(), rateLimitCode)
			RecordAccessError(r.Context(), NewAccessError("rate limited"))
			RecordAccessResponse(r.Context(), payload)
			gwlog.Warn(r.Context(), "rate limit rejected",
				gwlog.Field(gwlog.FieldAction, "rate_limit"),
				gwlog.Field(gwlog.FieldPath, r.URL.Path),
				gwlog.Field(gwlog.FieldRouteKey, m.routeKey),
			)
			WriteJSON(w, http.StatusTooManyRequests, payload)
			return
		}

		next(w, r)
	}
}
