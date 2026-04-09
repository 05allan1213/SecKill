package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func BitstormGetUserInfoByNameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserByNameRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewBitstormGetUserInfoByNameLogic(r.Context(), svcCtx)
		resp, err := l.BitstormGetUserInfoByName(&req)
		if err != nil {
			writeError(w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
