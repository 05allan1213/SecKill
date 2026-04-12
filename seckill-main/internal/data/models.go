package data

import "time"

type Goods struct {
	ID         int64 `gorm:"column:id"`
	GoodsNum   string
	GoodsName  string
	Price      float64
	PicUrl     string
	Seller     int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (p *Goods) TableName() string {
	return "t_goods"
}

type Order struct {
	ID         int64 `gorm:"column:id"`
	Seller     int64
	Buyer      int64
	OrderNum   string
	GoodsID    int64
	GoodsNum   string
	Price      float64
	Status     int
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (p *Order) TableName() string {
	return "t_order"
}

type Quota struct {
	ID         int64
	Num        int64
	GoodsID    int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (p *Quota) TableName() string {
	return "t_quota"
}

type SeckillMessage struct {
	TraceID string
	Goods   *Goods
	SecNum  string
	UserID  int64
	Num     int
}

type SecKillStatusEnum int

const (
	SK_STATUS_BEFORE_ORDER SecKillStatusEnum = 1
	SK_STATUS_BEFORE_PAY   SecKillStatusEnum = 2
	SK_STATUS_PAYED        SecKillStatusEnum = 3
	SK_STATUS_OOT          SecKillStatusEnum = 4
	SK_STATUS_CANCEL       SecKillStatusEnum = 5
	SK_STATUS_FAILED       SecKillStatusEnum = 6
)

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

func (p *SecKillRecord) TableName() string {
	return "t_seckill_record"
}

type SecKillStock struct {
	ID         int64 `gorm:"column:id"`
	GoodsID    int64
	Stock      int
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (p *SecKillStock) TableName() string {
	return "t_seckill_stock"
}

type PreSecKillRecord struct {
	SecNum     string
	UserID     int64
	GoodsID    int64
	GoodsNum   string
	OrderNum   string
	Price      float64
	Status     int
	Reason     string
	CreateTime time.Time
	ModifyTime time.Time
}

type UserQuota struct {
	ID         int64
	Num        int64
	KilledNum  int64
	UserID     int64
	GoodsID    int64
	CreateTime *time.Time `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (p *UserQuota) TableName() string {
	return "t_user_quota"
}

type SeckillAsyncResult struct {
	ID         int64       `gorm:"column:id"`
	SecNum     string      `gorm:"column:sec_num"`
	UserID     int64       `gorm:"column:user_id"`
	GoodsID    int64       `gorm:"column:goods_id"`
	GoodsNum   string      `gorm:"column:goods_num"`
	OrderNum   string      `gorm:"column:order_num"`
	Status     int         `gorm:"column:status"`
	Reason     string      `gorm:"column:reason"`
	Attempt    int         `gorm:"column:attempt"`
	LastError  string      `gorm:"column:last_error"`
	CreateTime *time.Time  `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time  `gorm:"column:modify_time;default:null"`
}

func (p *SeckillAsyncResult) TableName() string {
	return "t_seckill_async_result"
}
