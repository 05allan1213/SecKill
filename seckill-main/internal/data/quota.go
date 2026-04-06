package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
)

const (
	WITHOUT_QUOTA   int64 = -1
	WITHOUT_SETTING int64 = -2 //未设置
)

type quotaRepo struct {
	data *Data
}

// NewQuotaRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.QuotaRepo
func NewQuotaRepo(data *Data) biz.QuotaRepo {
	return &quotaRepo{
		data: data,
	}
}

func (r *quotaRepo) Save(ctx context.Context, data *biz.Data, g *biz.Quota) (*biz.Quota, error) {
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
//	@return *biz.Quota
//	@return error
func (r *quotaRepo) Update(ctx context.Context, data *biz.Data, g *biz.Quota) (*biz.Quota, error) {
	//err := db.Debug().Update(g).Error
	//return g, err
	return nil, nil
}

func (r *quotaRepo) FindByGoodsID(ctx context.Context, data *biz.Data, goodsID int64) (*biz.Quota, error) {
	var quota = new(biz.Quota)
	err := data.GetDB().Debug().WithContext(ctx).
		Where("goods_id = ?", goodsID).
		First(quota).Error
	return quota, err
}

func (r *quotaRepo) FindByIDWithCache(ctx context.Context, data *biz.Data, goodsID int64) (*biz.Quota, error) {
	cacheKey := fmt.Sprintf("quota:%d", goodsID)
	var quota = new(biz.Quota)
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
