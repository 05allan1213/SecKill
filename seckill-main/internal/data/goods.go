package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
	"time"
)

type GoodsRepo struct {
	data *Data
}

func NewGoodsRepo(data *Data) *GoodsRepo {
	return &GoodsRepo{
		data: data,
	}
}

func (r *GoodsRepo) Save(ctx context.Context, data *Data, g *Goods) (*Goods, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *GoodsRepo) Update(ctx context.Context, data *Data, g *Goods) (*Goods, error) {
	return nil, nil
}

func (r *GoodsRepo) FindByIDWithCache(ctx context.Context, data *Data, goodsID int64) (*Goods, error) {
	cacheKey := fmt.Sprintf("goodsinfo:%d", goodsID)
	var goods = new(Goods)
	rdbGoodsInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbGoodsInfo), goods)
		if err == nil {
			return goods, nil
		}
	}
	goods, err = r.FindByID(ctx, data, goodsID)
	if err != nil {
		return nil, err
	}
	goodsStr, _ := json.Marshal(goods)
	if goodsStr != nil && len(goodsStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(goodsStr), 10)
		if err != nil {
			log.InfoContextf(ctx, "set goods cacheKey err %s", err.Error())
		}
	}
	return goods, nil
}

func (r *GoodsRepo) FindByID(ctx context.Context, data *Data, goodsID int64) (*Goods, error) {
	var goods Goods
	err := data.GetDB().WithContext(ctx).Where("id = ?", goodsID).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}

func (r *GoodsRepo) FindByNum(ctx context.Context, data *Data, goodsNum string) (*Goods, error) {
	var goods Goods
	err := data.GetDB().WithContext(ctx).Where("goods_num = ?", goodsNum).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}

func (r *GoodsRepo) GetGoodsList(ctx context.Context, data *Data, offset int, limit int) ([]*Goods, error) {
	goodsList := make([]*Goods, 0)
	err := data.GetDB().WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&goodsList).Error
	if err != nil {
		return nil, err
	}
	return goodsList, err
}

func (r *GoodsRepo) GetGoodsInfoByNumWithCache(ctx context.Context, data *Data, goodsNum string) (*Goods, error) {
	cacheKey := fmt.Sprintf("goodsInfo:%s", goodsNum)
	var goods = new(Goods)
	rdbGoodsInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbGoodsInfo), goods)
		if err == nil {
			return goods, nil
		}
	}
	goods, err = r.FindByNum(ctx, data, goodsNum)
	if err != nil {
		return nil, err
	}
	goodsStr, _ := json.Marshal(goods)
	if len(goodsStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(goodsStr), 10*time.Second)
		if err != nil {
			log.InfoContextf(ctx, "set goods cacheKey err %s", err.Error())
		}
	}
	return goods, nil
}
