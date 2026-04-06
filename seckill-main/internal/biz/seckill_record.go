package biz

import (
	"context"
	"time"
)

type SecKillStatusEnum int

// 待支付；完成支付；超时未支付（关闭）;主动取消支付；
const (
	// 预购抢到名额，待生成订单
	SK_STATUS_BEFORE_ORDER SecKillStatusEnum = 1
	// 订单生成之后，待付款
	SK_STATUS_BEFORE_PAY SecKillStatusEnum = 2
	// 终态，已付款
	SK_STATUS_PAYED SecKillStatusEnum = 3
	// 终态，超时未付款
	SK_STATUS_OOT SecKillStatusEnum = 4
	// 终态，主动取消
	SK_STATUS_CANCEL SecKillStatusEnum = 5
)

// SecKillRecord is a SecKillRecord model.
type SecKillRecord struct {
	ID       int64 `gorm:"column:id"`
	SecNum   string
	UserID   int64
	GoodsID  int64
	OrderNum string
	Price    float64
	Status   int

	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

// TableName 表名
func (p *SecKillRecord) TableName() string {
	return "t_seckill_record"
}

// SecKillRecordRepo is a Greater repo.
type SecKillRecordRepo interface {
	Save(context.Context, *Data, *SecKillRecord) (*SecKillRecord, error)
	Update(context.Context, *Data, *SecKillRecord) (*SecKillRecord, error)
	FindByID(context.Context, *Data, int64) (*SecKillRecord, error)
	//DescStock(ctx context.Context, data *Data, goodsID int64, num int64) (int64, error)
	OutOfTime(context.Context, *Data, string) (int64, error)
}

// SecKillRecordUsecase is a SecKillRecord usecase.
type SecKillRecordUsecase struct {
	repo SecKillRecordRepo
}

// NewSecKillRecordUsecase new a SecKillRecord usecase.
func NewSecKillRecordUsecase(repo SecKillRecordRepo) *SecKillRecordUsecase {
	return &SecKillRecordUsecase{repo: repo}
}

// CreateSecKillRecord
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: creates a SecKillRecord, and returns the new SecKillRecord.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param g
//	@return *SecKillRecord
//	@return error
func (uc *SecKillRecordUsecase) CreateSecKillRecord(ctx context.Context, data *Data, g *SecKillRecord) (*SecKillRecord, error) {
	return uc.repo.Save(ctx, data, g)
}

// GetSecKillRecordInfo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: get  SecKillRecord, and returns new SecKillRecord.
//	@Receiver uc
//	@param ctx
//	@param data
//	@param userID
//	@return *SecKillRecord
//	@return error
func (uc *SecKillRecordUsecase) GetSecKillRecordInfo(ctx context.Context, data *Data, userID int64) (*SecKillRecord, error) {
	return uc.repo.FindByID(ctx, data, userID)
}

// SetOOTRecord
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver uc
//	@param ctx
//	@param data
//	@param orderID
//	@return int64
//	@return error
func (uc *SecKillRecordUsecase) SetOOTRecord(ctx context.Context, data *Data, orderID string) (int64, error) {
	return uc.repo.OutOfTime(ctx, data, orderID)
}
