package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
)

type secKillRecordRepo struct {
	data *Data
}

// NewSecKillRecordRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.SecKillRecordRepo
func NewSecKillRecordRepo(data *Data) biz.SecKillRecordRepo {
	return &secKillRecordRepo{
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
//	@return *biz.SecKillRecord
//	@return error
func (r *secKillRecordRepo) Save(ctx context.Context, data *biz.Data, g *biz.SecKillRecord) (*biz.SecKillRecord, error) {
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
//	@return *biz.SecKillRecord
//	@return error
func (r *secKillRecordRepo) Update(ctx context.Context, data *biz.Data, g *biz.SecKillRecord) (*biz.SecKillRecord, error) {
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
//	@return *biz.SecKillRecord
//	@return error
func (r *secKillRecordRepo) FindByIDWithCache(ctx context.Context, data *biz.Data,
	secKillID int64) (*biz.SecKillRecord, error) {
	cacheKey := fmt.Sprintf("secKillinfo:%d", secKillID)
	var secKill = new(biz.SecKillRecord)
	rdbSecKillRecordInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbSecKillRecordInfo), secKill)
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
//	@return *biz.SecKillRecord
//	@return error
func (r *secKillRecordRepo) FindByID(ctx context.Context, data *biz.Data, secKillID int64) (*biz.SecKillRecord, error) {
	var secKill biz.SecKillRecord
	err := data.GetDB().Debug().WithContext(ctx).Where("id = ?", secKillID).First(&secKill).Error
	if err != nil {
		return nil, err
	}
	return &secKill, nil
}

func (r *secKillRecordRepo) OutOfTime(ctx context.Context, data *biz.Data, orderID string) (int64, error) {
	db := data.GetDB()
	err := db.WithContext(ctx).Update("status", int(biz.SK_STATUS_OOT)).
		Where("order_id = ? and status = ?", orderID, int(biz.SK_STATUS_BEFORE_PAY)).Error
	return db.RowsAffected, err
}

// ListAll
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@return []*biz.SecKillRecord
//	@return error
func (r *secKillRecordRepo) ListAll(ctx context.Context, data *biz.Data) ([]*biz.SecKillRecord, error) {
	return nil, nil
}
