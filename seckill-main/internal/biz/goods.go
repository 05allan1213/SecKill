package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"time"
)

// Goods is a Goods model.
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

// TableName 表名
func (p *Goods) TableName() string {
	return "t_goods"
}

// GoodsRepo is a Greater repo.
type GoodsRepo interface {
	Save(context.Context, *Data, *Goods) (*Goods, error)
	Update(context.Context, *Data, *Goods) (*Goods, error)
	FindByID(context.Context, *Data, int64) (*Goods, error)
	FindByNum(context.Context, *Data, string) (*Goods, error)
	GetGoodsList(context.Context, *Data, int, int) ([]*Goods, error)
}

// GoodsUsecase is a Goods usecase.
type GoodsUsecase struct {
	repo GoodsRepo
}

// NewGoodsUsecase new a Goods usecase.
func NewGoodsUsecase(repo GoodsRepo) *GoodsUsecase {
	return &GoodsUsecase{repo: repo}
}

// CreateGoods
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: creates a Goods, and returns the new Goods.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param g
//	@return *Goods
//	@return error
func (uc *GoodsUsecase) CreateGoods(ctx context.Context, data *Data, g *Goods) (*Goods, error) {
	return uc.repo.Save(ctx, data, g)
}

// GetGoodsInfo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: get  Goods, and returns new Goods.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param userID
//	@return *Goods
//	@return error
func (uc *GoodsUsecase) GetGoodsInfo(ctx context.Context, data *Data, goodsID int64) (*Goods, error) {
	return uc.repo.FindByID(ctx, data, goodsID)
}

func (r *GoodsUsecase) GetGoodsInfoByNumWithCache(ctx context.Context, data *Data, goodsNum string) (*Goods, error) {
	cacheKey := fmt.Sprintf("goodsInfo:%d", goodsNum)
	var order = new(Goods)
	rdbOrderInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbOrderInfo), order)
		if err == nil {
			return order, nil
		}
	}
	order, err = r.GetGoodsInfoByNum(ctx, data, goodsNum)
	if err != nil {
		return nil, err
	}
	orderStr, _ := json.Marshal(order)
	if orderStr != nil && len(orderStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(orderStr), 10*time.Second)
		if err != nil {
			log.InfoContextf(ctx, "set order cacheKey err %s", err.Error())
		}
	}
	return order, nil
}

func (uc *GoodsUsecase) GetGoodsInfoByNum(ctx context.Context, data *Data, goodsNum string) (*Goods, error) {
	return uc.repo.FindByNum(ctx, data, goodsNum)
}

func (uc *GoodsUsecase) GetGoodsList(ctx context.Context, data *Data, offset int, limit int) ([]*Goods, error) {
	return uc.repo.GetGoodsList(ctx, data, offset, limit)
}
