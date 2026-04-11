package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BitofferHub/user/internal/log"
	"time"
)

type UserRepo struct {
	data *Data
}

func NewUserRepo(data *Data) *UserRepo {
	return &UserRepo{
		data: data,
	}
}

func (r *UserRepo) Save(ctx context.Context, data *Data, g *User) (*User, error) {
	err := data.GetDB().WithContext(ctx).Create(g).Error
	return g, err
}

func (r *UserRepo) Update(ctx context.Context, data *Data, g *User) (*User, error) {
	return nil, nil
}

func (r *UserRepo) FindByIDWithCache(ctx context.Context, data *Data, userID int64) (*User, error) {
	cacheKey := fmt.Sprintf("userinfo:%d", userID)
	var user = new(User)
	rdbUserInfo, exist, err := data.GetCache().Get(ctx, cacheKey)
	if err == nil && exist {
		err = json.Unmarshal([]byte(rdbUserInfo), user)
		if err == nil {
			return user, nil
		}
	}
	user, err = r.FindByID(ctx, data, userID)
	if err != nil {
		return nil, err
	}
	userStr, _ := json.Marshal(user)
	if len(userStr) != 0 {
		err = data.GetCache().Set(ctx, cacheKey, string(userStr), 10*time.Second)
		if err != nil {
			log.InfoContextf(ctx, "set user cacheKey err %s", err.Error())
		}
	}
	return user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, data *Data, userID int64) (*User, error) {
	var user User
	err := data.GetDB().WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByName(ctx context.Context, data *Data, userName string) (*User, error) {
	var user User
	err := data.GetDB().WithContext(ctx).Where("user_name = ?", userName).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) ListAll(ctx context.Context, data *Data) ([]*User, error) {
	return nil, nil
}
