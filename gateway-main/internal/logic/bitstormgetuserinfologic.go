package logic

import (
	"context"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"

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
	reply, err := fetchCurrentUser(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	return mapUserReply(reply), nil
}
