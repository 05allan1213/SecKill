package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
)

const (
	WITHOUT_QUOTA   int64 = -1
	WITHOUT_SETTING int64 = -2
)

type Quota struct {
	ID         int64
	Num        int64
	GoodsID    int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (q *Quota) TableName() string {
	return "t_quota"
}

type QuotaModel struct {
	store *Store
}

func NewQuotaModel(store *Store) *QuotaModel {
	return &QuotaModel{store: store}
}

func (m *QuotaModel) WithStore(store *Store) *QuotaModel {
	if store == nil {
		store = m.store
	}
	return &QuotaModel{store: store}
}

func (m *QuotaModel) CreateQuota(ctx context.Context, quota *Quota) (*Quota, error) {
	err := m.store.GetDB().WithContext(ctx).Create(quota).Error
	return quota, err
}

func (m *QuotaModel) FindByGoodsID(ctx context.Context, goodsID int64) (*Quota, error) {
	quota := new(Quota)
	err := m.store.GetDB().WithContext(ctx).
		Where("goods_id = ?", goodsID).
		First(quota).Error
	return quota, err
}

func (m *QuotaModel) FindByGoodsIDWithCache(ctx context.Context, goodsID int64) (*Quota, error) {
	cacheKey := fmt.Sprintf("quota:%d", goodsID)
	quota := new(Quota)
	rdbQuotaInfo, exist, err := m.store.GetCache().Get(ctx, cacheKey)
	if err == nil && exist && rdbQuotaInfo != "" {
		if err = json.Unmarshal([]byte(rdbQuotaInfo), quota); err == nil {
			return quota, nil
		}
	}

	quota, err = m.FindByGoodsID(ctx, goodsID)
	if err != nil {
		return nil, err
	}

	quotaStr, _ := json.Marshal(quota)
	if quotaStr != nil && len(quotaStr) != 0 {
		if err = m.store.GetCache().Set(ctx, cacheKey, string(quotaStr), 10); err != nil {
			log.InfoContextf(ctx, "set order cacheKey err %s", err.Error())
		}
	}
	return quota, nil
}
