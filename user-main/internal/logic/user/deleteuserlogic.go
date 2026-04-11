package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteUserLogic) DeleteUser(in *v1.DeleteUserRequest) (*v1.DeleteUserReply, error) {
	return &v1.DeleteUserReply{}, nil
}
