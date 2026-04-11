package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/svc"

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
	_, err := l.svcCtx.UserModel().CreateUser(l.ctx, newModelUserFromCreateRequest(in))
	if err != nil {
		return nil, err
	}
	return &v1.CreateUserReply{Message: "trytest"}, nil
}
