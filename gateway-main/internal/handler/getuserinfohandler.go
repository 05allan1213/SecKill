package handler

import (
	"net/http"

	"github.com/BitofferHub/gateway/internal/logic"
	"github.com/BitofferHub/gateway/internal/svc"
)

func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetUserInfoLogic(r.Context(), svcCtx)
		execute(w, r, l.GetUserInfo)
	}
}
