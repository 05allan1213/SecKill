package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
)

func BitstormGetSecKillInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBitstormGetSecKillInfoLogic(r.Context(), svcCtx)
		parseAndExecute[types.GetSecKillInfoRequest](w, r, l.BitstormGetSecKillInfo)
	}
}
