package logic

import (
	"context"
	"net/http"
	"time"

	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	userv1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, &middleware.HTTPError{Status: http.StatusBadRequest, Message: "missing Username or Password"}
	}

	reply, err := l.svcCtx.UserClient().GetUserByName(rpcContext(l.ctx), &userv1.GetUserByNameRequest{
		UserName: req.Username,
	})
	if err != nil {
		return nil, err
	}
	if reply == nil || reply.Data == nil || reply.Data.Pwd != req.Password {
		return nil, &middleware.HTTPError{Status: http.StatusUnauthorized, Message: "incorrect Username or Password"}
	}

	authConf := l.svcCtx.AuthConfig()
	token, expire, err := middleware.BuildToken(authConf.Secret, authConf.Timeout, reply.Data.UserID)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Code:   200,
		Token:  token,
		Expire: expire.Format(time.RFC3339),
	}, nil
}
