package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
	"gorm.io/gorm"
)

type SecKillStockRepo struct {
	data *Data
}

func NewSecKillStockRepo(data *Data) *SecKillStockRepo {
	return &SecKillStockRepo{
		data: data,
	}
}

func (r *SecKillStockRepo) Save(ctx context.Context, data *Data, g *SecKillStock) (*SecKillStock, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *SecKillStockRepo) Update(ctx context.Context, data *Data, g *SecKillStock) (*SecKillStock, error) {
	return nil, nil
}

func (r *SecKillStockRepo) FindByIDWithCache(ctx context.Context, data *Data, secKillID int64) (*SecKillStock, error) {
	cacheKey := fmt.Sprintf("secKillinfo:%d", secKillID)
	var secKill = new(SecKillStock)
	rdbSecKillStockInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbSecKillStockInfo), secKill)
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

func (r *SecKillStockRepo) FindByID(ctx context.Context, data *Data, secKillID int64) (*SecKillStock, error) {
	var secKill SecKillStock
	err := data.GetDB().WithContext(ctx).Where("id = ?", secKillID).First(&secKill).Error
	if err != nil {
		return nil, err
	}
	return &secKill, nil
}

func (r *SecKillStockRepo) DescStock(ctx context.Context, data *Data, goodsID int64, num int32) (int64, error) {
	var stock SecKillStock
	db := data.GetDB()
	db = db.WithContext(ctx).Table(stock.TableName()).
		Where("goods_id = ? and stock >= ?", goodsID, num).
		Update("stock", gorm.Expr("stock - ?", num))
	err := db.Error
	if err != nil {
		return 0, err
	}
	return db.RowsAffected, err
}

func (r *SecKillStockRepo) RebackStock(ctx context.Context, data *Data, goodsID int64, num int32) (int64, error) {
	var stock SecKillStock
	db := data.GetDB()
	err := db.Table(stock.TableName()).WithContext(ctx).Update("stock", gorm.Expr("stock + ?", num)).
		Where("goods_id= ?", goodsID).Error
	if err != nil {
		return 0, err
	}
	return db.RowsAffected, err
}

func (r *SecKillStockRepo) ListAll(ctx context.Context, data *Data) ([]*SecKillStock, error) {
	return nil, nil
}
