package userlogic

import (
	"context"

	"github.com/BitofferHub/user/internal/svc"
	v1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *v1.CreateUserRequest) (*v1.CreateUserReply, error) {
	return l.svcCtx.UserService.CreateUser(l.ctx, in)
}
