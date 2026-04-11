package logic

import (
	"context"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type BitstormSecKillV3Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBitstormSecKillV3Logic(ctx context.Context, svcCtx *svc.ServiceContext) *BitstormSecKillV3Logic {
	return &BitstormSecKillV3Logic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BitstormSecKillV3Logic) BitstormSecKillV3(req *types.SecKillRequest) (resp *types.SecKillV3Reply, err error) {
	return runSecKill(
		l.ctx,
		l.svcCtx,
		req,
		func(userID int64, req *types.SecKillRequest) *secproto.SecKillV3Request {
			return &secproto.SecKillV3Request{
				UserID:   userID,
				GoodsNum: req.GoodsNum,
				Num:      req.Num,
			}
		},
		func(ctx context.Context, req *secproto.SecKillV3Request) (*secproto.SecKillV3Reply, error) {
			return l.svcCtx.SeckillClient.SecKillV3(ctx, req)
		},
		mapSecKillV3Reply,
	)
}
