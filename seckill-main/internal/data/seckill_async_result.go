package data

import (
	"context"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
	"gorm.io/gorm"
)

type AsyncResultRepo interface {
	FindBySecNum(ctx context.Context, db *gorm.DB, secNum string) (*SeckillAsyncResult, error)
	UpsertPending(ctx context.Context, db *gorm.DB, secNum string, userID, goodsID int64, goodsNum string, attempt int) error
	UpsertSuccess(ctx context.Context, db *gorm.DB, secNum, orderNum string, status int) error
	UpsertFailure(ctx context.Context, db *gorm.DB, secNum, reason, lastError string, attempt int) error
}

type asyncResultRepo struct{}

func NewAsyncResultRepo() AsyncResultRepo {
	return &asyncResultRepo{}
}

func (r *asyncResultRepo) FindBySecNum(ctx context.Context, db *gorm.DB, secNum string) (*SeckillAsyncResult, error) {
	var result SeckillAsyncResult
	err := db.WithContext(ctx).Where("sec_num = ?", secNum).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *asyncResultRepo) UpsertPending(ctx context.Context, db *gorm.DB, secNum string, userID, goodsID int64, goodsNum string, attempt int) error {
	now := time.Now()
	result := SeckillAsyncResult{
		SecNum:    secNum,
		UserID:    userID,
		GoodsID:   goodsID,
		GoodsNum:  goodsNum,
		Status:    int(SK_STATUS_BEFORE_ORDER),
		Attempt:   attempt,
	}
	
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing SeckillAsyncResult
		err := tx.Where("sec_num = ?", secNum).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			result.CreateTime = &now
			result.ModifyTime = &now
			return tx.Create(&result).Error
		}
		if err != nil {
			return err
		}
		updates := map[string]interface{}{
			"attempt":     attempt,
			"modify_time": now,
		}
		if existing.Status == int(SK_STATUS_BEFORE_ORDER) {
			updates["status"] = result.Status
		}
		return tx.Model(&SeckillAsyncResult{}).
			Where("sec_num = ?", secNum).
			Updates(updates).Error
	})
}

func (r *asyncResultRepo) UpsertSuccess(ctx context.Context, db *gorm.DB, secNum, orderNum string, status int) error {
	now := time.Now()
	return db.WithContext(ctx).Model(&SeckillAsyncResult{}).
		Where("sec_num = ?", secNum).
		Updates(map[string]interface{}{
			"status":      status,
			"order_num":   orderNum,
			"modify_time": now,
		}).Error
}

func (r *asyncResultRepo) UpsertFailure(ctx context.Context, db *gorm.DB, secNum, reason, lastError string, attempt int) error {
	now := time.Now()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing SeckillAsyncResult
		err := tx.Where("sec_num = ?", secNum).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			log.Warn(ctx, "async result not found for failure upsert",
				log.Field(log.FieldSecNum, secNum),
			)
			result := SeckillAsyncResult{
				SecNum:     secNum,
				Status:     int(SK_STATUS_FAILED),
				Reason:     reason,
				LastError:  lastError,
				Attempt:    attempt,
				CreateTime: &now,
				ModifyTime: &now,
			}
			return tx.Create(&result).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&SeckillAsyncResult{}).
			Where("sec_num = ?", secNum).
			Updates(map[string]interface{}{
				"status":      int(SK_STATUS_FAILED),
				"reason":      reason,
				"last_error":  lastError,
				"attempt":     attempt,
				"modify_time": now,
			}).Error
	})
}
