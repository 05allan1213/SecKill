package biz

import (
	"context"
	"time"
)

// Quota is a Quota model.
type Quota struct {
	ID         int64
	Num        int64
	GoodsID    int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *Quota) TableName() string {
	return "t_quota"
}

// QuotaRepo is a Greater repo.
type QuotaRepo interface {
	Save(context.Context, *Data, *Quota) (*Quota, error)
	Update(context.Context, *Data, *Quota) (*Quota, error)
	FindByGoodsID(context.Context, *Data, int64) (*Quota, error)
}

// QuotaUsecase is a Quota usecase.
type QuotaUsecase struct {
	repo QuotaRepo
}

// NewQuotaUsecase new a Quota usecase.
func NewQuotaUsecase(repo QuotaRepo) *QuotaUsecase {
	return &QuotaUsecase{repo: repo}
}

// CreateQuota
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: creates a Quota, and returns the new Quota.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param g
//	@return *Quota
//	@return error
func (uc *QuotaUsecase) CreateQuota(ctx context.Context, data *Data, g *Quota) (*Quota, error) {
	return uc.repo.Save(ctx, data, g)
}

// GetQuotaInfo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: get  Quota, and returns new Quota.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param userID
//	@return *Quota
//	@return error
func (uc *QuotaUsecase) FindByGoodsID(ctx context.Context, data *Data, goodsID int64) (*Quota, error) {
	return uc.repo.FindByGoodsID(ctx, data, goodsID)
}
