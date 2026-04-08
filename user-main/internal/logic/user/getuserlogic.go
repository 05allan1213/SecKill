package userlogic

import (
	"context"

	"github.com/BitofferHub/user/internal/svc"
	v1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *v1.GetUserRequest) (*v1.GetUserReply, error) {
	return l.svcCtx.UserService.GetUser(l.ctx, in)
}
