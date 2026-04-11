package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
)

func BitstormSecKillV3Handler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBitstormSecKillV3Logic(r.Context(), svcCtx)
		parseAndExecute[types.SecKillRequest](w, r, l.BitstormSecKillV3)
	}
}
