package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
)

type goodsRepo struct {
	data *Data
}

// NewGoodsRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.GoodsRepo
func NewGoodsRepo(data *Data) biz.GoodsRepo {
	return &goodsRepo{
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
//	@return *biz.Goods
//	@return error
func (r *goodsRepo) Save(ctx context.Context, data *biz.Data, g *biz.Goods) (*biz.Goods, error) {
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
//	@return *biz.Goods
//	@return error
func (r *goodsRepo) Update(ctx context.Context, data *biz.Data, g *biz.Goods) (*biz.Goods, error) {
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
//	@param goodsID
//	@return *biz.Goods
//	@return error
func (r *goodsRepo) FindByIDWithCache(ctx context.Context, data *biz.Data,
	goodsID int64) (*biz.Goods, error) {
	cacheKey := fmt.Sprintf("goodsinfo:%d", goodsID)
	var goods = new(biz.Goods)
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

// FindByID
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@param goodsID
//	@return *biz.Goods
//	@return error
func (r *goodsRepo) FindByID(ctx context.Context, data *biz.Data, goodsID int64) (*biz.Goods, error) {
	var goods biz.Goods
	err := data.GetDB().Debug().WithContext(ctx).Where("id = ?", goodsID).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}

func (r *goodsRepo) FindByNum(ctx context.Context, data *biz.Data, goodsNum string) (*biz.Goods, error) {
	var goods biz.Goods
	fmt.Println("data is before")
	fmt.Printf("db is %s\n", GetJsonFmtStr(data.GetDB()))
	fmt.Println("goodsNum ", goodsNum)
	err := data.GetDB().Debug().WithContext(ctx).Where("goods_num = ?", goodsNum).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}

// GetJsonFmtStr func
func GetJsonFmtStr(data interface{}) string {
	resp, _ := json.Marshal(data)
	respStr := string(resp)
	if respStr == "" {
		respStr = fmt.Sprintf("%+v", data)
	}
	return respStr
}

// ListAll
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@return []*biz.Goods
//	@return error
func (r *goodsRepo) GetGoodsList(ctx context.Context, data *biz.Data, offset int, limit int) ([]*biz.Goods, error) {

	goodsList := make([]*biz.Goods, 0)
	err := data.GetDB().Debug().WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Find(&goodsList).Error
	if err != nil {
		return nil, err
	}
	return goodsList, err
}
