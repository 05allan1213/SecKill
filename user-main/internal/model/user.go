package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitofferHub/user/internal/log"
)

const userCacheTTLSeconds = 10

type User struct {
	UserID     int64 `gorm:"column:id;primaryKey;autoIncrement"`
	UserName   string
	Pwd        string
	Sex        int
	Age        int
	Email      string
	Contact    string
	Mobile     string
	IdCard     string
	CreateTime time.Time  `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (u *User) TableName() string {
	return "t_user_info"
}

type UserModel struct {
	store *Store
}

func NewUserModel(store *Store) *UserModel {
	return &UserModel{store: store}
}

func (m *UserModel) CreateUser(ctx context.Context, user *User) (*User, error) {
	err := m.store.DB().WithContext(ctx).Create(user).Error
	return user, err
}

func (m *UserModel) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	var user User
	err := m.store.DB().WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *UserModel) GetUserByIDWithCache(ctx context.Context, userID int64) (*User, error) {
	cacheKey := userCacheKey(userID)
	var user User

	rdbUserInfo, exist, err := m.store.Cache().Get(ctx, cacheKey)
	if err == nil && exist {
		if err := json.Unmarshal([]byte(rdbUserInfo), &user); err == nil {
			return &user, nil
		}
	}

	userInfo, err := m.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userStr, _ := json.Marshal(userInfo)
	if len(userStr) > 0 {
		if err := m.store.Cache().Set(ctx, cacheKey, string(userStr), userCacheTTLSeconds); err != nil {
			log.InfoContextf(ctx, "set user cacheKey err %s", err.Error())
		}
	}
	return userInfo, nil
}

func (m *UserModel) GetUserByName(ctx context.Context, userName string) (*User, error) {
	var user User
	err := m.store.DB().WithContext(ctx).Where("user_name = ?", userName).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func userCacheKey(userID int64) string {
	return fmt.Sprintf("userinfo:%d", userID)
}
