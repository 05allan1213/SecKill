package data

import (
	"context"
	"github.com/BitofferHub/seckill/internal/biz"
	"gorm.io/gorm"
)

type userUserQuotaRepo struct {
	data *Data
}

// NewUserQuotaRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.UserQuotaRepo
func NewUserQuotaRepo(data *Data) biz.UserQuotaRepo {
	return &userUserQuotaRepo{
		data: data,
	}
}

func (r *userUserQuotaRepo) Save(ctx context.Context, data *biz.Data, g *biz.UserQuota) (*biz.UserQuota, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
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
//	@return *biz.UserQuota
//	@return error
func (r *userUserQuotaRepo) Update(ctx context.Context, data *biz.Data, g *biz.UserQuota) (*biz.UserQuota, error) {
	return nil, nil
}

func (r *userUserQuotaRepo) FindByGoodsID(ctx context.Context, data *biz.Data, goodsID int64) (*biz.UserQuota, error) {
	var userUserQuota = new(biz.UserQuota)
	err := data.GetDB().WithContext(ctx).First(userUserQuota).Error
	return userUserQuota, err
}

func (r *userUserQuotaRepo) FindUserGoodsQuota(ctx context.Context, data *biz.Data, userID int64, goodsID int64) (*biz.UserQuota, error) {
	var userUserQuota = new(biz.UserQuota)
	err := data.GetDB().WithContext(ctx).
		Where("user_id = ? and goods_id = ?", userID, goodsID).
		First(userUserQuota).Error
	return userUserQuota, err
}

func (r *userUserQuotaRepo) IncrKilledNum(ctx context.Context, data *biz.Data,
	userID int64, goodsID int64, num int64) (int64, error) {
	var uq biz.UserQuota
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
