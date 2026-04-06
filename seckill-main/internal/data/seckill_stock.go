package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
	"gorm.io/gorm"
)

type secKillStockRepo struct {
	data *Data
}

// NewSecKillStockRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.SecKillStockRepo
func NewSecKillStockRepo(data *Data) biz.SecKillStockRepo {
	return &secKillStockRepo{
		data: data,
	}
}

// Save
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@param g
//	@return *biz.SecKillStock
//	@return error
func (r *secKillStockRepo) Save(ctx context.Context, data *biz.Data, g *biz.SecKillStock) (*biz.SecKillStock, error) {
	err := data.GetDB().Debug().WithContext(ctx).Create(g).Error
	return g, err
}

// Update
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@param g
//	@return *biz.SecKillStock
//	@return error
func (r *secKillStockRepo) Update(ctx context.Context, data *biz.Data, g *biz.SecKillStock) (*biz.SecKillStock, error) {
	//err := db.Debug().Update(g).Error
	//return g, err
	return nil, nil
}

// FindByIDWithCache
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@param secKillID
//	@return *biz.SecKillStock
//	@return error
func (r *secKillStockRepo) FindByIDWithCache(ctx context.Context, data *biz.Data,
	secKillID int64) (*biz.SecKillStock, error) {
	cacheKey := fmt.Sprintf("secKillinfo:%d", secKillID)
	var secKill = new(biz.SecKillStock)
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

// FindByID
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@param secKillID
//	@return *biz.SecKillStock
//	@return error
func (r *secKillStockRepo) FindByID(ctx context.Context, data *biz.Data, secKillID int64) (*biz.SecKillStock, error) {
	var secKill biz.SecKillStock
	err := data.GetDB().Debug().WithContext(ctx).Where("id = ?", secKillID).First(&secKill).Error
	if err != nil {
		return nil, err
	}
	return &secKill, nil
}

func (r *secKillStockRepo) DescStock(ctx context.Context, data *biz.Data, goodsID int64, num int32) (int64, error) {
	var stock biz.SecKillStock
	db := data.GetDB()
	db = db.Debug().WithContext(ctx).Table(stock.TableName()).
		Where("goods_id = ? and stock >= ?", goodsID, num).
		Update("stock", gorm.Expr("stock - ?", num))
	err := db.Error
	if err != nil {
		return 0, err
	}
	return db.RowsAffected, err
}

func (r *secKillStockRepo) RebackStock(ctx context.Context, data *biz.Data, goodsID int64, num int32) (int64, error) {
	var stock biz.SecKillStock
	db := data.GetDB()
	err := db.Table(stock.TableName()).WithContext(ctx).Update("stock", gorm.Expr("stock + ?", num)).
		Where("goods_id= ?", goodsID).Error
	if err != nil {
		return 0, err
	}
	return db.RowsAffected, err
}

// ListAll
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@return []*biz.SecKillStock
//	@return error
func (r *secKillStockRepo) ListAll(ctx context.Context, data *biz.Data) ([]*biz.SecKillStock, error) {
	return nil, nil
}
