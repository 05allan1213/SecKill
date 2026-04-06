package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
)

type orderRepo struct {
	data *Data
}

// NewOrderRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.OrderRepo
func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{
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
//	@return *biz.Order
//	@return error
func (r *orderRepo) Save(ctx context.Context, data *biz.Data, g *biz.Order) (*biz.Order, error) {
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
//	@return *biz.Order
//	@return error
func (r *orderRepo) Update(ctx context.Context, data *biz.Data, g *biz.Order) (*biz.Order, error) {
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
//	@param orderID
//	@return *biz.Order
//	@return error
func (r *orderRepo) FindByIDWithCache(ctx context.Context, data *biz.Data,
	orderID int64) (*biz.Order, error) {
	cacheKey := fmt.Sprintf("orderinfo:%d", orderID)
	var order = new(biz.Order)
	rdbOrderInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbOrderInfo), order)
		if err == nil {
			return order, nil
		}
	}
	order, err = r.FindByID(ctx, data, orderID)
	if err != nil {
		return nil, err
	}
	orderStr, _ := json.Marshal(order)
	if orderStr != nil && len(orderStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(orderStr), 10)
		if err != nil {
			log.InfoContextf(ctx, "set order cacheKey err %s", err.Error())
		}
	}
	return order, nil
}

// FindByID
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@param orderID
//	@return *biz.Order
//	@return error
func (r *orderRepo) FindByID(ctx context.Context, data *biz.Data, orderID int64) (*biz.Order, error) {
	var order biz.Order
	err := data.GetDB().Debug().WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepo) FindByNum(ctx context.Context, data *biz.Data, orderNum int64) (*biz.Order, error) {
	var order biz.Order
	err := data.GetDB().Debug().WithContext(ctx).Where("order_num = ?", orderNum).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ListAll
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver r
//	@param ctx
//	@param data
//	@return []*biz.Order
//	@return error
func (r *orderRepo) ListAll(ctx context.Context, data *biz.Data) ([]*biz.Order, error) {
	return nil, nil
}
