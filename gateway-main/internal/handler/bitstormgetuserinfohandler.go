package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func BitstormGetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBitstormGetUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.BitstormGetUserInfo()
		if err != nil {
			writeError(w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
