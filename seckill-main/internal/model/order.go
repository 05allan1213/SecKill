package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
)

type Order struct {
	ID         int64 `gorm:"column:id"`
	Seller     int64
	Buyer      int64
	OrderNum   string
	GoodsID    int64
	GoodsNum   string
	Price      float64
	Status     int
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (o *Order) TableName() string {
	return "t_order"
}

type OrderModel struct {
	store *Store
}

func NewOrderModel(store *Store) *OrderModel {
	return &OrderModel{store: store}
}

func (m *OrderModel) WithStore(store *Store) *OrderModel {
	if store == nil {
		store = m.store
	}
	return &OrderModel{store: store}
}

func (m *OrderModel) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	err := m.store.GetDB().WithContext(ctx).Create(order).Error
	return order, err
}

func (m *OrderModel) GetOrderInfo(ctx context.Context, orderID int64) (*Order, error) {
	var order Order
	err := m.store.GetDB().WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (m *OrderModel) GetOrderInfoWithCache(ctx context.Context, orderID int64) (*Order, error) {
	cacheKey := fmt.Sprintf("orderinfo:%d", orderID)
	order := new(Order)
	rdbOrderInfo, exist, err := m.store.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		if err = json.Unmarshal([]byte(rdbOrderInfo), order); err == nil {
			return order, nil
		}
	}

	order, err = m.GetOrderInfo(ctx, orderID)
	if err != nil {
		return nil, err
	}

	orderStr, _ := json.Marshal(order)
	if orderStr != nil && len(orderStr) != 0 {
		if err = m.store.GetCache().Set(ctx, cacheKey, string(orderStr), 10); err != nil {
			log.InfoContextf(ctx, "set order cacheKey err %s", err.Error())
		}
	}
	return order, nil
}
