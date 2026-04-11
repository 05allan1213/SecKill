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
	reply := new(pb.GetGoodsListReply)
	goodsList, err := l.svcCtx.GoodsRepo.GetGoodsList(l.ctx, l.svcCtx.Data, int(req.Offset), int(req.Limit))
	if err != nil {
		log.ErrorContextf(l.ctx, "get secinfo by secnum err %s\n", err.Error())
		return nil, err
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
