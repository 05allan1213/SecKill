package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/seckill/internal/log"
)

type OrderRepo struct {
	data *Data
}

func NewOrderRepo(data *Data) *OrderRepo {
	return &OrderRepo{
		data: data,
	}
}

func (r *OrderRepo) Save(ctx context.Context, data *Data, g *Order) (*Order, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *OrderRepo) Update(ctx context.Context, data *Data, g *Order) (*Order, error) {
	return nil, nil
}

func (r *OrderRepo) FindByIDWithCache(ctx context.Context, data *Data, orderID int64) (*Order, error) {
	cacheKey := fmt.Sprintf("orderinfo:%d", orderID)
	var order = new(Order)
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

func (r *OrderRepo) FindByID(ctx context.Context, data *Data, orderID int64) (*Order, error) {
	var order Order
	err := data.GetDB().WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) FindByNum(ctx context.Context, data *Data, orderNum int64) (*Order, error) {
	var order Order
	err := data.GetDB().WithContext(ctx).Where("order_num = ?", orderNum).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) ListAll(ctx context.Context, data *Data) ([]*Order, error) {
	return nil, nil
}
