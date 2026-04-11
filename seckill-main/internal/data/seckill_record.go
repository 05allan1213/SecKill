package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
)

type SecKillRecordRepo struct {
	data *Data
}

func NewSecKillRecordRepo(data *Data) *SecKillRecordRepo {
	return &SecKillRecordRepo{
		data: data,
	}
}

func (r *SecKillRecordRepo) Save(ctx context.Context, data *Data, g *SecKillRecord) (*SecKillRecord, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *SecKillRecordRepo) Update(ctx context.Context, data *Data, g *SecKillRecord) (*SecKillRecord, error) {
	return nil, nil
}

func (r *SecKillRecordRepo) FindByIDWithCache(ctx context.Context, data *Data, secKillID int64) (*SecKillRecord, error) {
	cacheKey := fmt.Sprintf("secKillinfo:%d", secKillID)
	var secKill = new(SecKillRecord)
	rdbSecKillRecordInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbSecKillRecordInfo), secKill)
		if err == nil {
			return secKill, nil
		}
	}
	secKill, err = r.FindByID(ctx, data, secKillID)
	if err != nil {
		return nil, err
	}
	secKillStr, _ := json.Marshal(secKill)
	if secKillStr != nil && len(secKillStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(secKillStr), 10)
		if err != nil {
			log.InfoContextf(ctx, "set secKill cacheKey err %s", err.Error())
		}
	}
	return secKill, nil
}

func (r *SecKillRecordRepo) FindByID(ctx context.Context, data *Data, secKillID int64) (*SecKillRecord, error) {
	var secKill SecKillRecord
	err := data.GetDB().WithContext(ctx).Where("id = ?", secKillID).First(&secKill).Error
	if err != nil {
		return nil, err
	}
	return &secKill, nil
}

func (r *SecKillRecordRepo) OutOfTime(ctx context.Context, data *Data, orderID string) (int64, error) {
	db := data.GetDB()
	err := db.WithContext(ctx).Update("status", int(SK_STATUS_OOT)).
		Where("order_id = ? and status = ?", orderID, int(SK_STATUS_BEFORE_PAY)).Error
	return db.RowsAffected, err
}

func (r *SecKillRecordRepo) ListAll(ctx context.Context, data *Data) ([]*SecKillRecord, error) {
	return nil, nil
}
