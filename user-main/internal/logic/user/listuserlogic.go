package userlogic

import (
	"context"

	"github.com/BitofferHub/user/internal/svc"
	v1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserLogic {
	return &ListUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserLogic) ListUser(in *v1.ListUserRequest) (*v1.ListUserReply, error) {
	return &v1.ListUserReply{}, nil
}
