package userlogic

import (
	"context"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/data"
	"github.com/BitofferHub/user/internal/log"
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
	log.Info(l.ctx, "create user request",
		log.Field(log.FieldAction, "user.create"),
		log.Field("userName", in.UserName),
	)
	_, err := l.svcCtx.UserRepo.Save(l.ctx, l.svcCtx.Data, &data.User{
		UserName: in.UserName,
		Pwd:      in.Pwd,
		Sex:      int(in.Sex),
		Age:      int(in.Age),
		Email:    in.Email,
		Contact:  in.Contact,
		Mobile:   in.Mobile,
		IdCard:   in.IdCard,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateUserReply{Message: "trytest"}, nil
}
