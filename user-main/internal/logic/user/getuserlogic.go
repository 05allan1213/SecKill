package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/log"
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
	l.ctx = log.WithFields(l.ctx,
		log.Field(log.FieldAction, "user.get"),
		log.Field(log.FieldUserID, in.UserID),
	)
	userInfo, err := l.svcCtx.UserRepo.FindByID(l.ctx, l.svcCtx.Data, in.UserID)
	if err != nil {
		log.Error(l.ctx, "find user by id failed",
			log.Field(log.FieldAction, "user.get"),
			log.Field(log.FieldError, err.Error()),
		)
		return nil, userRepoError(err)
	}
	return &v1.GetUserReply{
		Code:    0,
		Message: "success",
		Data: &v1.GetUserReplyData{
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
