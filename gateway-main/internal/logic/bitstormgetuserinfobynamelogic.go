package logic

import (
	"context"
	"net/http"

	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	userv1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type BitstormGetUserInfoByNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBitstormGetUserInfoByNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BitstormGetUserInfoByNameLogic {
	return &BitstormGetUserInfoByNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BitstormGetUserInfoByNameLogic) BitstormGetUserInfoByName(req *types.GetUserByNameRequest) (resp *types.GetUserReply, err error) {
	userName := req.UserName
	if userName == "" {
		userName = req.UserNameAlt
	}
	if userName == "" {
		return nil, &middleware.HTTPError{Status: http.StatusBadRequest, Message: "user_name is required"}
	}

	reply, err := l.svcCtx.UserClient.GetUserByName(rpcContext(l.ctx), &userv1.GetUserByNameRequest{
		UserName: userName,
	})
	if err != nil {
		return nil, err
	}

	return mapUserByNameReply(reply), nil
}
