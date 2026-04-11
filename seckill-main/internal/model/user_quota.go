package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserQuota struct {
	ID         int64
	Num        int64
	KilledNum  int64
	UserID     int64
	GoodsID    int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (u *UserQuota) TableName() string {
	return "t_user_quota"
}

type UserQuotaModel struct {
	store *Store
}

func NewUserQuotaModel(store *Store) *UserQuotaModel {
	return &UserQuotaModel{store: store}
}

func (m *UserQuotaModel) WithStore(store *Store) *UserQuotaModel {
	if store == nil {
		store = m.store
	}
	return &UserQuotaModel{store: store}
}

func (m *UserQuotaModel) CreateUserQuota(ctx context.Context, quota *UserQuota) (*UserQuota, error) {
	err := m.store.GetDB().WithContext(ctx).Create(quota).Error
	return quota, err
}

func (m *UserQuotaModel) FindByGoodsID(ctx context.Context, goodsID int64) (*UserQuota, error) {
	userQuota := new(UserQuota)
	err := m.store.GetDB().WithContext(ctx).First(userQuota).Error
	return userQuota, err
}

func (m *UserQuotaModel) FindUserGoodsQuota(ctx context.Context, userID int64, goodsID int64) (*UserQuota, error) {
	userQuota := new(UserQuota)
	err := m.store.GetDB().WithContext(ctx).
		Where("user_id = ? and goods_id = ?", userID, goodsID).
		First(userQuota).Error
	return userQuota, err
}

func (m *UserQuotaModel) IncrKilledNum(ctx context.Context, userID int64, goodsID int64, num int64) (int64, error) {
	var uq UserQuota
	db := m.store.GetDB()
	db = db.WithContext(ctx).Table(uq.TableName()).
		Where("user_id = ? and goods_id = ?", userID, goodsID).
		Update("killed_num", gorm.Expr("killed_num + ?", num))
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
