package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
)

func BitstormGetUserInfoByNameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBitstormGetUserInfoByNameLogic(r.Context(), svcCtx)
		parseAndExecute[types.GetUserByNameRequest](w, r, l.BitstormGetUserInfoByName)
	}
}
