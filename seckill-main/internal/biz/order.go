package biz

import (
	"context"
	"time"
)

// Order is a Order model.
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

// TableName 表名
func (p *Order) TableName() string {
	return "t_order"
}

// OrderRepo is a Greater repo.
type OrderRepo interface {
	Save(context.Context, *Data, *Order) (*Order, error)
	Update(context.Context, *Data, *Order) (*Order, error)
	FindByID(context.Context, *Data, int64) (*Order, error)
}

// OrderUsecase is a Order usecase.
type OrderUsecase struct {
	repo OrderRepo
}

// NewOrderUsecase new a Order usecase.
func NewOrderUsecase(repo OrderRepo) *OrderUsecase {
	return &OrderUsecase{repo: repo}
}

// CreateOrder
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: creates a Order, and returns the new Order.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param g
//	@return *Order
//	@return error
func (uc *OrderUsecase) CreateOrder(ctx context.Context, data *Data, g *Order) (*Order, error) {
	return uc.repo.Save(ctx, data, g)
}

// GetOrderInfo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: get  Order, and returns new Order.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param userID
//	@return *Order
//	@return error
func (uc *OrderUsecase) GetOrderInfo(ctx context.Context, data *Data, userID int64) (*Order, error) {
	return uc.repo.FindByID(ctx, data, userID)
}
