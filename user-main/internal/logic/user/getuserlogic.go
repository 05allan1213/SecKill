package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/svc"

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
	userInfo, err := l.svcCtx.UserModel().GetUserByID(l.ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.GetUserReply{
		Code:    0,
		Message: "success",
		Data:    newUserReplyData(userInfo, false),
	}, nil
}
