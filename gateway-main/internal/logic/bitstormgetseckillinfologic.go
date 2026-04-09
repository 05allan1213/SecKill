package logic

import (
	"context"
	"net/http"

	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type BitstormGetSecKillInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBitstormGetSecKillInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BitstormGetSecKillInfoLogic {
	return &BitstormGetSecKillInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BitstormGetSecKillInfoLogic) BitstormGetSecKillInfo(req *types.GetSecKillInfoRequest) (resp *types.GetSecKillInfoReply, err error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.SecNum == "" {
		return nil, &middleware.HTTPError{Status: http.StatusBadRequest, Message: "sec_num is required"}
	}

	reply, err := l.svcCtx.SeckillClient.GetSecKillInfo(rpcContext(l.ctx), &secproto.GetSecKillInfoRequest{
		UserID: userID,
		SecNum: req.SecNum,
	})
	if err != nil {
		return nil, err
	}

	return mapGetSecKillInfoReply(reply), nil
}
