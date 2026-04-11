package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
)

type SecKillStatusEnum int

const (
	SK_STATUS_BEFORE_ORDER SecKillStatusEnum = 1
	SK_STATUS_BEFORE_PAY   SecKillStatusEnum = 2
	SK_STATUS_PAYED        SecKillStatusEnum = 3
	SK_STATUS_OOT          SecKillStatusEnum = 4
	SK_STATUS_CANCEL       SecKillStatusEnum = 5
)

type SecKillRecord struct {
	ID       int64 `gorm:"column:id"`
	SecNum   string
	UserID   int64
	GoodsID  int64
	OrderNum string
	Price    float64
	Status   int

	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (r *SecKillRecord) TableName() string {
	return "t_seckill_record"
}

type SecKillRecordModel struct {
	store *Store
}

func NewSecKillRecordModel(store *Store) *SecKillRecordModel {
	return &SecKillRecordModel{store: store}
}

func (m *SecKillRecordModel) WithStore(store *Store) *SecKillRecordModel {
	if store == nil {
		store = m.store
	}
	return &SecKillRecordModel{store: store}
}

func (m *SecKillRecordModel) CreateSecKillRecord(ctx context.Context, record *SecKillRecord) (*SecKillRecord, error) {
	err := m.store.GetDB().WithContext(ctx).Create(record).Error
	return record, err
}

func (m *SecKillRecordModel) GetSecKillRecordInfo(ctx context.Context, secKillID int64) (*SecKillRecord, error) {
	var record SecKillRecord
	err := m.store.GetDB().WithContext(ctx).Where("id = ?", secKillID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (m *SecKillRecordModel) GetSecKillRecordInfoWithCache(ctx context.Context, secKillID int64) (*SecKillRecord, error) {
	cacheKey := fmt.Sprintf("secKillinfo:%d", secKillID)
	record := new(SecKillRecord)
	rdbSecKillRecordInfo, exist, err := m.store.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		if err = json.Unmarshal([]byte(rdbSecKillRecordInfo), record); err == nil {
			return record, nil
		}
	}

	record, err = m.GetSecKillRecordInfo(ctx, secKillID)
	if err != nil {
		return nil, err
	}

	recordStr, _ := json.Marshal(record)
	if recordStr != nil && len(recordStr) != 0 {
		if err = m.store.GetCache().Set(ctx, cacheKey, string(recordStr), 10); err != nil {
			log.InfoContextf(ctx, "set secKill cacheKey err %s", err.Error())
		}
	}
	return record, nil
}

func (m *SecKillRecordModel) SetOOTRecord(ctx context.Context, orderID string) (int64, error) {
	db := m.store.GetDB()
	err := db.WithContext(ctx).Where("order_id = ? and status = ?", orderID, int(SK_STATUS_BEFORE_PAY)).
		Update("status", int(SK_STATUS_OOT)).Error
	return db.RowsAffected, err
}
