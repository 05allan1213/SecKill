package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
	"gorm.io/gorm"
)

type SecKillStock struct {
	ID         int64 `gorm:"column:id"`
	GoodsID    int64
	Stock      int
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (s *SecKillStock) TableName() string {
	return "t_seckill_stock"
}

type SecKillStockModel struct {
	store *Store
}

func NewSecKillStockModel(store *Store) *SecKillStockModel {
	return &SecKillStockModel{store: store}
}

func (m *SecKillStockModel) WithStore(store *Store) *SecKillStockModel {
	if store == nil {
		store = m.store
	}
	return &SecKillStockModel{store: store}
}

func (m *SecKillStockModel) CreateSecKillStock(ctx context.Context, stock *SecKillStock) (*SecKillStock, error) {
	err := m.store.GetDB().WithContext(ctx).Create(stock).Error
	return stock, err
}

func (m *SecKillStockModel) GetSecKillStockInfo(ctx context.Context, stockID int64) (*SecKillStock, error) {
	var stock SecKillStock
	err := m.store.GetDB().WithContext(ctx).Where("id = ?", stockID).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (m *SecKillStockModel) GetSecKillStockInfoWithCache(ctx context.Context, stockID int64) (*SecKillStock, error) {
	cacheKey := fmt.Sprintf("secKillinfo:%d", stockID)
	stock := new(SecKillStock)
	rdbStockInfo, exist, err := m.store.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		if err = json.Unmarshal([]byte(rdbStockInfo), stock); err == nil {
			return stock, nil
		}
	}

	stock, err = m.GetSecKillStockInfo(ctx, stockID)
	if err != nil {
		return nil, err
	}

	stockStr, _ := json.Marshal(stock)
	if stockStr != nil && len(stockStr) != 0 {
		if err = m.store.GetCache().Set(ctx, cacheKey, string(stockStr), 10); err != nil {
			log.InfoContextf(ctx, "set secKill cacheKey err %s", err.Error())
		}
	}
	return stock, nil
}

func (m *SecKillStockModel) DescStock(ctx context.Context, goodsID int64, num int32) (int64, error) {
	var stock SecKillStock
	db := m.store.GetDB().WithContext(ctx).Table(stock.TableName()).
		Where("goods_id = ? and stock >= ?", goodsID, num).
		Update("stock", gorm.Expr("stock - ?", num))
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (m *SecKillStockModel) RebackStock(ctx context.Context, goodsID int64, num int32) (int64, error) {
	var stock SecKillStock
	db := m.store.GetDB().WithContext(ctx).Table(stock.TableName()).
		Where("goods_id = ?", goodsID).
		Update("stock", gorm.Expr("stock + ?", num))
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
