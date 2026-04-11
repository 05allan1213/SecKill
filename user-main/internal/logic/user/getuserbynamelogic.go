package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/log"
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
	l.ctx = log.WithFields(l.ctx,
		log.Field(log.FieldAction, "user.get_by_name"),
		log.Field("userName", in.UserName),
	)
	userInfo, err := l.svcCtx.UserRepo.FindByName(l.ctx, l.svcCtx.Data, in.UserName)
	if err != nil {
		log.Error(l.ctx, "find user by name failed",
			log.Field(log.FieldAction, "user.get_by_name"),
			log.Field(log.FieldError, err.Error()),
		)
		return nil, userRepoError(err)
	}
	return &v1.GetUserByNameReply{
		Code:    0,
		Message: "success",
		Data: &v1.GetUserReplyData{
			UserID:   userInfo.UserID,
			UserName: userInfo.UserName,
			Pwd:      userInfo.Pwd,
			Sex:      int32(userInfo.Sex),
			Age:      int32(userInfo.Age),
			Email:    userInfo.Email,
			Contact:  userInfo.Contact,
			Mobile:   userInfo.Mobile,
			IdCard:   userInfo.IdCard,
		},
	}, nil
}
