package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/redis/go-redis/v9"
)

func TestPreSecKillStockRepo_PreDescStockAndSetFailed(t *testing.T) {
	ctx := context.Background()
	dt := newRedisTestData(t, 9)
	repo := NewPreSecKillStockRepo(dt)

	userID := int64(9101)
	goodsID := int64(9201)
	num := int32(2)
	secNum := fmt.Sprintf("sec-test-%d", time.Now().UnixNano())
	stockKey := fmt.Sprintf("SK:Stock:%d", goodsID)
	userSecNumKey := fmt.Sprintf("SK:UserGoodsSecNum:%d:%d", userID, goodsID)
	userKilledKey := fmt.Sprintf("SK:UserSecKilledNum:%d:%d", userID, goodsID)

	clearRedisKeys(t, dt, stockKey, userSecNumKey, userKilledKey, secNum)
	t.Cleanup(func() {
		clearRedisKeys(t, dt, stockKey, userSecNumKey, userKilledKey, secNum)
	})

	if err := dt.GetCache().Set(ctx, stockKey, "5", 5*time.Minute); err != nil {
		t.Fatalf("set stock key: %v", err)
	}

	record := &PreSecKillRecord{
		SecNum:     secNum,
		UserID:     userID,
		GoodsID:    goodsID,
		GoodsNum:   "abc123",
		Status:     int(SK_STATUS_BEFORE_ORDER),
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}

	gotSecNum, err := repo.PreDescStock(ctx, dt, userID, goodsID, num, secNum, record)
	if err != nil {
		t.Fatalf("pre desc stock: %v", err)
	}
	if gotSecNum != secNum {
		t.Fatalf("unexpected secNum: got %q want %q", gotSecNum, secNum)
	}
	if got := mustGetRedisValue(t, dt, stockKey); got != "3" {
		t.Fatalf("unexpected stock after pre desc: got %q want 3", got)
	}
	if got := mustGetRedisValue(t, dt, userSecNumKey); got != secNum {
		t.Fatalf("unexpected user secNum key: got %q want %q", got, secNum)
	}
	if got := mustGetRedisValue(t, dt, userKilledKey); got != "2" {
		t.Fatalf("unexpected user killed num: got %q want 2", got)
	}

	record.Status = int(SK_STATUS_FAILED)
	record.Reason = "mq send failed"
	record.ModifyTime = time.Now()
	if _, err := repo.SetFailedInPreSecKill(ctx, dt, userID, goodsID, num, secNum, record); err != nil {
		t.Fatalf("set failed in pre seckill: %v", err)
	}

	if got := mustGetRedisValue(t, dt, stockKey); got != "5" {
		t.Fatalf("unexpected stock after rollback: got %q want 5", got)
	}
	if got := mustGetRedisValue(t, dt, userSecNumKey); got != "" {
		t.Fatalf("unexpected user secNum after rollback: got %q want empty", got)
	}
	if got := mustGetRedisValue(t, dt, userKilledKey); got != "0" {
		t.Fatalf("unexpected user killed num after rollback: got %q want 0", got)
	}

	failedRecord, err := repo.GetSecKillInfo(ctx, dt, secNum)
	if err != nil {
		t.Fatalf("get seckill info: %v", err)
	}
	if failedRecord.Status != int(SK_STATUS_FAILED) {
		t.Fatalf("unexpected failed status: got %d want %d", failedRecord.Status, SK_STATUS_FAILED)
	}
	if failedRecord.Reason != "mq send failed" {
		t.Fatalf("unexpected failure reason: got %q", failedRecord.Reason)
	}
}

func TestPreSecKillStockRepo_PreDescStockDuplicate(t *testing.T) {
	ctx := context.Background()
	dt := newRedisTestData(t, 9)
	repo := NewPreSecKillStockRepo(dt)

	userID := int64(9102)
	goodsID := int64(9202)
	num := int32(1)
	firstSecNum := fmt.Sprintf("sec-first-%d", time.Now().UnixNano())
	secondSecNum := fmt.Sprintf("sec-second-%d", time.Now().UnixNano())
	stockKey := fmt.Sprintf("SK:Stock:%d", goodsID)
	userSecNumKey := fmt.Sprintf("SK:UserGoodsSecNum:%d:%d", userID, goodsID)
	userKilledKey := fmt.Sprintf("SK:UserSecKilledNum:%d:%d", userID, goodsID)

	clearRedisKeys(t, dt, stockKey, userSecNumKey, userKilledKey, firstSecNum, secondSecNum)
	t.Cleanup(func() {
		clearRedisKeys(t, dt, stockKey, userSecNumKey, userKilledKey, firstSecNum, secondSecNum)
	})

	if err := dt.GetCache().Set(ctx, stockKey, "5", 5*time.Minute); err != nil {
		t.Fatalf("set stock key: %v", err)
	}

	firstRecord := &PreSecKillRecord{
		SecNum:     firstSecNum,
		UserID:     userID,
		GoodsID:    goodsID,
		GoodsNum:   "abc123",
		Status:     int(SK_STATUS_BEFORE_ORDER),
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	if _, err := repo.PreDescStock(ctx, dt, userID, goodsID, num, firstSecNum, firstRecord); err != nil {
		t.Fatalf("first pre desc stock: %v", err)
	}

	secondRecord := &PreSecKillRecord{
		SecNum:     secondSecNum,
		UserID:     userID,
		GoodsID:    goodsID,
		GoodsNum:   "abc123",
		Status:     int(SK_STATUS_BEFORE_ORDER),
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	gotSecNum, err := repo.PreDescStock(ctx, dt, userID, goodsID, num, secondSecNum, secondRecord)
	if err == nil {
		t.Fatal("expected duplicate seckill error")
	}
	if err != SecKillErrSecKilling {
		t.Fatalf("unexpected duplicate error: %v", err)
	}
	if gotSecNum != firstSecNum {
		t.Fatalf("unexpected existing secNum: got %q want %q", gotSecNum, firstSecNum)
	}
}

func newRedisTestData(t *testing.T, db int) *Data {
	t.Helper()

	rawClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       db,
	})
	if err := rawClient.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}

	cache.Init(
		cache.WithAddr("127.0.0.1:6379"),
		cache.WithPassWord("123456"),
		cache.WithDB(db),
		cache.WithPoolSize(10),
	)
	return &Data{rdb: cache.GetRedisCli()}
}

func clearRedisKeys(t *testing.T, dt *Data, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if err := dt.GetCache().Delete(context.Background(), key); err != nil {
			t.Fatalf("delete redis key %s: %v", key, err)
		}
	}
}

func mustGetRedisValue(t *testing.T, dt *Data, key string) string {
	t.Helper()
	value, _, err := dt.GetCache().Get(context.Background(), key)
	if err != nil {
		t.Fatalf("get redis key %s: %v", key, err)
	}
	return value
}
