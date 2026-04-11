//go:build integration

package seckill

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/model"
	"github.com/BitofferHub/seckill/internal/svc"
)

func TestSeckillLogicIntegration(t *testing.T) {
	cfg, err := config.Load("../../../etc/seckill.yaml")
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	svcCtx := svc.NewServiceContext(cfg)
	t.Cleanup(svcCtx.Close)

	ctx := context.Background()
	seedSeckillFixtures(t, svcCtx)

	goodsReply, err := NewGetGoodsListLogic(ctx, svcCtx).GetGoodsList(&pb.GetGoodsListRequest{
		UserID: 1,
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("get goods list failed: %v", err)
	}
	if goodsReply.Data == nil || len(goodsReply.Data.GoodsList) == 0 {
		t.Fatalf("unexpected goods list reply: %#v", goodsReply)
	}

	v1Reply, err := NewSecKillV1Logic(ctx, svcCtx).SecKillV1(&pb.SecKillV1Request{
		UserID:   910001,
		GoodsNum: "abc123",
		Num:      1,
	})
	if err != nil {
		t.Fatalf("seckill v1 failed: %v", err)
	}
	if v1Reply.Code != 0 || v1Reply.Data == nil || v1Reply.Data.OrderNum == "" {
		t.Fatalf("unexpected seckill v1 reply: %#v", v1Reply)
	}

	goods, err := svcCtx.GoodsModel().GetGoodsInfoByNumWithCache(ctx, "abc123")
	if err != nil {
		t.Fatalf("load goods failed: %v", err)
	}

	secNum := fmt.Sprintf("itest-sec-%d", time.Now().UnixNano())
	record := newPreSeckillRecord(secNum, 910002, goods)
	if _, err := svcCtx.PreStockModel().PreDescStock(ctx, 910002, goods.ID, 1, secNum, &record); err != nil {
		t.Fatalf("pre desc stock failed: %v", err)
	}

	message, err := json.Marshal(&model.SeckillMessage{
		TraceID: "itest-trace",
		Goods:   goods,
		SecNum:  secNum,
		UserID:  910002,
		Num:     1,
	})
	if err != nil {
		t.Fatalf("marshal message failed: %v", err)
	}
	if err := handleConsumedMessage(ctx, svcCtx, message); err != nil {
		t.Fatalf("handle consumed message failed: %v", err)
	}

	infoReply, err := NewGetSecKillInfoLogic(ctx, svcCtx).GetSecKillInfo(&pb.GetSecKillInfoRequest{
		UserID: 910002,
		SecNum: secNum,
	})
	if err != nil {
		t.Fatalf("get sec kill info failed: %v", err)
	}
	if infoReply.Data == nil || infoReply.Data.SecNum != secNum || infoReply.Data.Status != int32(model.SK_STATUS_BEFORE_PAY) {
		t.Fatalf("unexpected sec kill info reply: %#v", infoReply)
	}
}

func seedSeckillFixtures(t *testing.T, svcCtx *svc.ServiceContext) {
	t.Helper()

	db := svcCtx.Store().GetDB()
	cache := svcCtx.Store().GetCache()
	ctx := context.Background()

	statements := []string{
		"INSERT INTO t_quota(goods_id, num) VALUES (1, 10) ON DUPLICATE KEY UPDATE num = VALUES(num)",
		"UPDATE t_seckill_stock SET stock = 20 WHERE goods_id = 1",
		"DELETE FROM t_user_quota WHERE user_id IN (910001, 910002)",
		"DELETE FROM t_seckill_record WHERE user_id IN (910001, 910002)",
		"DELETE FROM t_order WHERE buyer IN (910001, 910002)",
	}
	for _, statement := range statements {
		if err := db.WithContext(ctx).Exec(statement).Error; err != nil {
			t.Fatalf("seed db failed for %q: %v", statement, err)
		}
	}

	if err := cache.Set(ctx, "SK:Limit1", "10", 10*time.Minute); err != nil {
		t.Fatalf("seed limit cache failed: %v", err)
	}
	if err := cache.Set(ctx, "SK:Stock:1", "20", 10*time.Minute); err != nil {
		t.Fatalf("seed stock cache failed: %v", err)
	}
	for _, key := range []string{
		"SK:UserGoodsSecNum:910001:1",
		"SK:UserGoodsSecNum:910002:1",
		"SK:UserSecKilledNum:910001:1",
		"SK:UserSecKilledNum:910002:1",
	} {
		if err := cache.Delete(ctx, key); err != nil {
			t.Fatalf("cleanup cache key %q failed: %v", key, err)
		}
	}
}
