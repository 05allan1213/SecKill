package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
)

type Goods struct {
	ID         int64 `gorm:"column:id"`
	GoodsNum   string
	GoodsName  string
	Price      float64
	PicUrl     string
	Seller     int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (g *Goods) TableName() string {
	return "t_goods"
}

type GoodsModel struct {
	store *Store
}

func NewGoodsModel(store *Store) *GoodsModel {
	return &GoodsModel{store: store}
}

func (m *GoodsModel) WithStore(store *Store) *GoodsModel {
	if store == nil {
		store = m.store
	}
	return &GoodsModel{store: store}
}

func (m *GoodsModel) CreateGoods(ctx context.Context, g *Goods) (*Goods, error) {
	err := m.store.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (m *GoodsModel) GetGoodsInfo(ctx context.Context, goodsID int64) (*Goods, error) {
	var goods Goods
	err := m.store.GetDB().WithContext(ctx).Where("id = ?", goodsID).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}

func (m *GoodsModel) GetGoodsInfoByNum(ctx context.Context, goodsNum string) (*Goods, error) {
	var goods Goods
	err := m.store.GetDB().WithContext(ctx).Where("goods_num = ?", goodsNum).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}

func (m *GoodsModel) GetGoodsInfoByNumWithCache(ctx context.Context, goodsNum string) (*Goods, error) {
	cacheKey := fmt.Sprintf("goodsInfo:%s", goodsNum)
	goods := new(Goods)
	rdbGoodsInfo, exist, err := m.store.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		if err = json.Unmarshal([]byte(rdbGoodsInfo), goods); err == nil {
			return goods, nil
		}
	}

	goods, err = m.GetGoodsInfoByNum(ctx, goodsNum)
	if err != nil {
		return nil, err
	}

	goodsStr, _ := json.Marshal(goods)
	if goodsStr != nil && len(goodsStr) != 0 {
		if err = m.store.GetCache().Set(ctx, cacheKey, string(goodsStr), 10*time.Second); err != nil {
			log.InfoContextf(ctx, "set order cacheKey err %s", err.Error())
		}
	}
	return goods, nil
}

func (m *GoodsModel) GetGoodsList(ctx context.Context, offset int, limit int) ([]*Goods, error) {
	goodsList := make([]*Goods, 0)
	err := m.store.GetDB().WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&goodsList).Error
	if err != nil {
		return nil, err
	}
	return goodsList, nil
}
