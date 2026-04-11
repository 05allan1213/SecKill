package middleware

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/svc"
)

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
		limiter := m.svcCtx.Limiter()
		if limiter == nil {
			next(w, r)
			return
		}

		result, err := limiter.Allow(r.Context(), m.routeKey)
		if err != nil || !result.IsAllowed {
			WriteJSON(w, http.StatusOK, "ok")
			return
		}

		next(w, r)
	}
}
