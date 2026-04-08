package userlogic

import (
	"context"

	"github.com/BitofferHub/user/internal/svc"
	v1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByNameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByNameLogic {
	return &GetUserByNameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByNameLogic) GetUserByName(in *v1.GetUserByNameRequest) (*v1.GetUserByNameReply, error) {
	return l.svcCtx.UserService.GetUserByName(l.ctx, in)
}
