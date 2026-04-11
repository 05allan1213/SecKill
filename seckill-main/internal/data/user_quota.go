package data

import (
	"context"
	"gorm.io/gorm"
)

type UserQuotaRepo struct {
	data *Data
}

func NewUserQuotaRepo(data *Data) *UserQuotaRepo {
	return &UserQuotaRepo{
		data: data,
	}
}

func (r *UserQuotaRepo) Save(ctx context.Context, data *Data, g *UserQuota) (*UserQuota, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *UserQuotaRepo) Update(ctx context.Context, data *Data, g *UserQuota) (*UserQuota, error) {
	return nil, nil
}

func (r *UserQuotaRepo) FindByGoodsID(ctx context.Context, data *Data, goodsID int64) (*UserQuota, error) {
	var userUserQuota = new(UserQuota)
	err := data.GetDB().WithContext(ctx).First(userUserQuota).Error
	return userUserQuota, err
}

func (r *UserQuotaRepo) FindUserGoodsQuota(ctx context.Context, data *Data, userID int64, goodsID int64) (*UserQuota, error) {
	var userUserQuota = new(UserQuota)
	err := data.GetDB().WithContext(ctx).
		Where("user_id = ? and goods_id = ?", userID, goodsID).
		First(userUserQuota).Error
	return userUserQuota, err
}

func (r *UserQuotaRepo) IncrKilledNum(ctx context.Context, data *Data,
	userID int64, goodsID int64, num int64) (int64, error) {
	var uq UserQuota
	db := data.GetDB()
	db = db.WithContext(ctx).Table(uq.TableName()).
		Where("user_id = ? and goods_id = ?", userID, goodsID).
		Update("killed_num", gorm.Expr("killed_num + ?", num))
	err := db.Error
	if err != nil {
		return 0, err
	}
	return db.RowsAffected, err
}
