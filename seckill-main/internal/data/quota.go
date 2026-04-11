package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
)

const (
	WITHOUT_QUOTA   int64 = -1
	WITHOUT_SETTING int64 = -2 //未设置
)

type QuotaRepo struct {
	data *Data
}

func NewQuotaRepo(data *Data) *QuotaRepo {
	return &QuotaRepo{
		data: data,
	}
}

func (r *QuotaRepo) Save(ctx context.Context, data *Data, g *Quota) (*Quota, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *QuotaRepo) Update(ctx context.Context, data *Data, g *Quota) (*Quota, error) {
	return nil, nil
}

func (r *QuotaRepo) FindByGoodsID(ctx context.Context, data *Data, goodsID int64) (*Quota, error) {
	var quota = new(Quota)
	err := data.GetDB().WithContext(ctx).
		Where("goods_id = ?", goodsID).
		First(quota).Error
	return quota, err
}

func (r *QuotaRepo) FindByIDWithCache(ctx context.Context, data *Data, goodsID int64) (*Quota, error) {
	cacheKey := fmt.Sprintf("quota:%d", goodsID)
	var quota = new(Quota)
	rdbQuotaInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist && rdbQuotaInfo != "" {
		err = json.Unmarshal([]byte(rdbQuotaInfo), quota)
		if err == nil {
			return quota, nil
		}
	}
	quota, err = r.FindByGoodsID(ctx, data, goodsID)
	if err != nil {
		return nil, err
	}
	quotaStr, _ := json.Marshal(quota)
	if quotaStr != nil && len(quotaStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(quotaStr), 10)
		if err != nil {
			log.InfoContextf(ctx, "set order cacheKey err %s", err.Error())
		}
	}
	return quota, nil
}
