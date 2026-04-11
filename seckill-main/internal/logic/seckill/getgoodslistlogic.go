package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGoodsListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGoodsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGoodsListLogic {
	return &GetGoodsListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetGoodsListLogic) GetGoodsList(req *pb.GetGoodsListRequest) (*pb.GetGoodsListReply, error) {
	l.ctx = log.WithAction(l.ctx, "GetGoodsList")
	reply := new(pb.GetGoodsListReply)
	goodsList, err := l.svcCtx.GoodsRepo.GetGoodsList(l.ctx, l.svcCtx.Data, int(req.Offset), int(req.Limit))
	if err != nil {
		log.Error(l.ctx, "get goods list failed",
			log.Field(log.FieldAction, "GetGoodsList"),
			log.Field("offset", req.Offset),
			log.Field("limit", req.Limit),
			log.Field(log.FieldError, err.Error()),
		)
		return nil, dependencyUnavailableError("goods list unavailable")
	}

	reply.Data = &pb.GetGoodsListReplyData{
		GoodsList: make([]*pb.GoodInfo, 0, len(goodsList)),
	}
	for _, goods := range goodsList {
		info := new(pb.GoodInfo)
		convertDataGoodsToPbGoods(goods, info)
		reply.Data.GoodsList = append(reply.Data.GoodsList, info)
	}
	return reply, nil
}
