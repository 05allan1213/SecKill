package biz

import (
	"context"
	"time"
)

// UserQuota is a UserQuota model.
type UserQuota struct {
	ID         int64
	Num        int64
	KilledNum  int64
	UserID     int64
	GoodsID    int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *UserQuota) TableName() string {
	return "t_user_quota"
}

// UserQuotaRepo is a Greater repo.
type UserQuotaRepo interface {
	Save(context.Context, *Data, *UserQuota) (*UserQuota, error)
	Update(context.Context, *Data, *UserQuota) (*UserQuota, error)
	FindByGoodsID(context.Context, *Data, int64) (*UserQuota, error)
	FindUserGoodsQuota(context.Context, *Data, int64, int64) (*UserQuota, error)
	IncrKilledNum(context.Context, *Data, int64, int64, int64) (int64, error)
}

// UserQuotaUsecase is a UserQuota usecase.
type UserQuotaUsecase struct {
	repo UserQuotaRepo
}

// NewUserQuotaUsecase new a UserQuota usecase.
func NewUserQuotaUsecase(repo UserQuotaRepo) *UserQuotaUsecase {
	return &UserQuotaUsecase{repo: repo}
}

// CreateUserQuota
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: creates a UserQuota, and returns the new UserQuota.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param g
//	@return *UserQuota
//	@return error
func (uc *UserQuotaUsecase) CreateUserQuota(ctx context.Context, data *Data, g *UserQuota) (*UserQuota, error) {
	return uc.repo.Save(ctx, data, g)
}

// GetUserQuotaInfo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: get  UserQuota, and returns new UserQuota.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param userID
//	@return *UserQuota
//	@return error
func (uc *UserQuotaUsecase) FindByGoodsID(ctx context.Context, data *Data, goodsID int64) (*UserQuota, error) {
	return uc.repo.FindByGoodsID(ctx, data, goodsID)
}

func (uc *UserQuotaUsecase) FindUserGoodsQuota(ctx context.Context, data *Data, userID int64, goodsID int64) (*UserQuota, error) {
	return uc.repo.FindUserGoodsQuota(ctx, data, userID, goodsID)
}

func (uc *UserQuotaUsecase) IncrKilledNum(ctx context.Context, data *Data, userID int64, goodsID int64, num int64) (int64, error) {
	return uc.repo.IncrKilledNum(ctx, data, userID, goodsID, num)

}
