package biz

import (
	"context"
	"time"
)

// SecKillStock is a SecKillStock model.
type SecKillStock struct {
	ID         int64 `gorm:"column:id"`
	GoodsID    int64
	Stock      int
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *SecKillStock) TableName() string {
	return "t_seckill_stock"
}

// SecKillStockRepo is a Greater repo.
type SecKillStockRepo interface {
	Save(context.Context, *Data, *SecKillStock) (*SecKillStock, error)
	Update(context.Context, *Data, *SecKillStock) (*SecKillStock, error)
	FindByID(context.Context, *Data, int64) (*SecKillStock, error)
	DescStock(context.Context, *Data, int64, int32) (int64, error)
	RebackStock(context.Context, *Data, int64, int32) (int64, error)
	//	Try(ctx context.Context, data *Data, a int) error
	//DescStock(ctx context.Context, data *Data, goodsID int64, num int64) (int64, error)
}

// SecKillStockUsecase is a SecKillStock usecase.
type SecKillStockUsecase struct {
	repo SecKillStockRepo
}

// NewSecKillStockUsecase new a SecKillStock usecase.
func NewSecKillStockUsecase(repo SecKillStockRepo) *SecKillStockUsecase {
	return &SecKillStockUsecase{repo: repo}
}

type PreSecKillRecord struct {
	SecNum     string //key
	UserID     int64
	GoodsID    int64
	OrderNum   string
	Price      float64
	Status     int
	CreateTime time.Time
	ModifyTime time.Time
}

// CreateSecKillStock
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: creates a SecKillStock, and returns the new SecKillStock.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param g
//	@return *SecKillStock
//	@return error
func (uc *SecKillStockUsecase) CreateSecKillStock(ctx context.Context, data *Data, g *SecKillStock) (*SecKillStock, error) {
	return uc.repo.Save(ctx, data, g)
}

// GetSecKillStockInfo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: get  SecKillStock, and returns new SecKillStock.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param userID
//	@return *SecKillStock
//	@return error
func (uc *SecKillStockUsecase) GetSecKillStockInfo(ctx context.Context, data *Data, userID int64) (*SecKillStock, error) {
	return uc.repo.FindByID(ctx, data, userID)
}

func (uc *SecKillStockUsecase) DescStock(ctx context.Context, data *Data, goodsID int64, num int32) (int64, error) {
	return uc.repo.DescStock(ctx, data, goodsID, num)
}
