package logic

import (
	"context"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	userv1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type BitstormGetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBitstormGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BitstormGetUserInfoLogic {
	return &BitstormGetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BitstormGetUserInfoLogic) BitstormGetUserInfo() (resp *types.GetUserReply, err error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	reply, err := l.svcCtx.UserClient.GetUser(rpcContext(l.ctx), &userv1.GetUserRequest{UserID: userID})
	if err != nil {
		return nil, err
	}

	return mapUserReply(reply), nil
}
