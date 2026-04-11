package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/svc"

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
	userInfo, err := l.svcCtx.UserModel().GetUserByName(l.ctx, in.UserName)
	if err != nil {
		return nil, err
	}
	return &v1.GetUserByNameReply{
		Code:    0,
		Message: "success",
		Data:    newUserReplyData(userInfo, true),
	}, nil
}
