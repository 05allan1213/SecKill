package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetOrderInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderInfoLogic {
	return &GetOrderInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderInfoLogic) GetOrderInfo(req *pb.GetOrderInfoRequest) (*pb.GetOrderInfoReply, error) {
	return nil, status.Error(codes.Unimplemented, "GetOrderInfo is not implemented")
}
