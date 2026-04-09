package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	seckilllogic "github.com/BitofferHub/seckill/internal/logic/seckill"
	"github.com/BitofferHub/seckill/internal/svc"
)

type SecKillServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedSecKillServer
}

func NewSecKillServer(svcCtx *svc.ServiceContext) *SecKillServer {
	return &SecKillServer{svcCtx: svcCtx}
}

func (s *SecKillServer) SecKillV1(ctx context.Context, req *pb.SecKillV1Request) (*pb.SecKillV1Reply, error) {
	return seckilllogic.NewSecKillV1Logic(ctx, s.svcCtx).SecKillV1(req)
}

func (s *SecKillServer) SecKillV2(ctx context.Context, req *pb.SecKillV2Request) (*pb.SecKillV2Reply, error) {
	return seckilllogic.NewSecKillV2Logic(ctx, s.svcCtx).SecKillV2(req)
}

func (s *SecKillServer) SecKillV3(ctx context.Context, req *pb.SecKillV3Request) (*pb.SecKillV3Reply, error) {
	return seckilllogic.NewSecKillV3Logic(ctx, s.svcCtx).SecKillV3(req)
}

func (s *SecKillServer) GetGoodsList(ctx context.Context, req *pb.GetGoodsListRequest) (*pb.GetGoodsListReply, error) {
	return seckilllogic.NewGetGoodsListLogic(ctx, s.svcCtx).GetGoodsList(req)
}

func (s *SecKillServer) GetSecKillInfo(ctx context.Context, req *pb.GetSecKillInfoRequest) (*pb.GetSecKillInfoReply, error) {
	return seckilllogic.NewGetSecKillInfoLogic(ctx, s.svcCtx).GetSecKillInfo(req)
}

func (s *SecKillServer) GetOrderList(ctx context.Context, req *pb.GetOrderListRequest) (*pb.GetOrderListReply, error) {
	return seckilllogic.NewGetOrderListLogic(ctx, s.svcCtx).GetOrderList(req)
}

func (s *SecKillServer) GetOrderInfo(ctx context.Context, req *pb.GetOrderInfoRequest) (*pb.GetOrderInfoReply, error) {
	return seckilllogic.NewGetOrderInfoLogic(ctx, s.svcCtx).GetOrderInfo(req)
}
