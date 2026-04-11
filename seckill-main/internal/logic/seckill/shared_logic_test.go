package seckill

import (
	"context"
	"testing"

	"github.com/BitofferHub/pkg/constant"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/model"
)

func TestGetErrMsg(t *testing.T) {
	if got := GetErrMsg(ERR_GOODS_STOCK_NOT_ENOUGH); got != "商品库存不足" {
		t.Fatalf("unexpected known error message: %s", got)
	}
	if got := GetErrMsg(999999); got != "unknown error code 999999" {
		t.Fatalf("unexpected unknown error message: %s", got)
	}
}

func TestTraceIDFromContext(t *testing.T) {
	if got := traceIDFromContext(nil); got != "" {
		t.Fatalf("expected empty trace id for nil ctx, got %q", got)
	}

	ctx := context.WithValue(context.Background(), constant.TraceID, "trace-123")
	if got := traceIDFromContext(ctx); got != "trace-123" {
		t.Fatalf("unexpected trace id: %q", got)
	}
}

func TestConvertBizGoodsToPbGoods(t *testing.T) {
	src := &model.Goods{
		GoodsNum:  "GOODS001",
		GoodsName: "demo",
		Price:     99.9,
		PicUrl:    "https://example.test/goods.png",
		Seller:    42,
	}
	dst := new(pb.GoodInfo)

	convertBizGoodsToPbGoods(src, dst)

	if dst.GoodsNum != src.GoodsNum || dst.GoodsName != src.GoodsName {
		t.Fatalf("basic fields not copied: %+v", dst)
	}
	if dst.Seller != src.Seller || dst.PicUrl != src.PicUrl {
		t.Fatalf("seller/pic fields not copied: %+v", dst)
	}
	if dst.Price != float32(src.Price) {
		t.Fatalf("price not converted: %+v", dst)
	}
}
