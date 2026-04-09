package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func BitstormSecKillV1Handler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SecKillRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewBitstormSecKillV1Logic(r.Context(), svcCtx)
		resp, err := l.BitstormSecKillV1(&req)
		if err != nil {
			writeError(w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
