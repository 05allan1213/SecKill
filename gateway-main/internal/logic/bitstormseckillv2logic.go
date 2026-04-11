package logic

import (
	"context"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/internal/types"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type BitstormSecKillV2Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBitstormSecKillV2Logic(ctx context.Context, svcCtx *svc.ServiceContext) *BitstormSecKillV2Logic {
	return &BitstormSecKillV2Logic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BitstormSecKillV2Logic) BitstormSecKillV2(req *types.SecKillRequest) (resp *types.SecKillV2Reply, err error) {
	return runSecKill(
		l.ctx,
		l.svcCtx,
		req,
		func(userID int64, req *types.SecKillRequest) *secproto.SecKillV2Request {
			return &secproto.SecKillV2Request{
				UserID:   userID,
				GoodsNum: req.GoodsNum,
				Num:      req.Num,
			}
		},
		func(ctx context.Context, req *secproto.SecKillV2Request) (*secproto.SecKillV2Reply, error) {
			return l.svcCtx.SeckillClient.SecKillV2(ctx, req)
		},
		mapSecKillV2Reply,
	)
}
