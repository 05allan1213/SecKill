package logic

import (
	"context"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type BitstormSecKillV1Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBitstormSecKillV1Logic(ctx context.Context, svcCtx *svc.ServiceContext) *BitstormSecKillV1Logic {
	return &BitstormSecKillV1Logic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BitstormSecKillV1Logic) BitstormSecKillV1(req *types.SecKillRequest) (resp *types.SecKillV1Reply, err error) {
	return runSecKill(
		l.ctx,
		l.svcCtx,
		req,
		func(userID int64, req *types.SecKillRequest) *secproto.SecKillV1Request {
			return &secproto.SecKillV1Request{
				UserID:   userID,
				GoodsNum: req.GoodsNum,
				Num:      req.Num,
			}
		},
		func(ctx context.Context, req *secproto.SecKillV1Request) (*secproto.SecKillV1Reply, error) {
			return l.svcCtx.SeckillClient.SecKillV1(ctx, req)
		},
		mapSecKillV1Reply,
	)
}
