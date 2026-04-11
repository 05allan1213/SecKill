package logic

import (
	"context"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.WelcomeResponse, err error) {
	reply, err := fetchCurrentUser(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	return &types.WelcomeResponse{
		Welcome: reply.GetData().GetUserName(),
	}, nil
}
