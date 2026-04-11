package data

type Repositories struct {
	StockRepo     *SecKillStockRepo
	PreStockRepo  *PreSecKillStockRepo
	RecordRepo    *SecKillRecordRepo
	GoodsRepo     *GoodsRepo
	OrderRepo     *OrderRepo
	MessageRepo   *SecKillMsgRepo
	QuotaRepo     *QuotaRepo
	UserQuotaRepo *UserQuotaRepo
}

func NewRepositories(dataLayer *Data) *Repositories {
	return &Repositories{
		StockRepo:     NewSecKillStockRepo(dataLayer),
		PreStockRepo:  NewPreSecKillStockRepo(dataLayer),
		RecordRepo:    NewSecKillRecordRepo(dataLayer),
		GoodsRepo:     NewGoodsRepo(dataLayer),
		OrderRepo:     NewOrderRepo(dataLayer),
		MessageRepo:   NewSecKillMsgRepo(dataLayer),
		QuotaRepo:     NewQuotaRepo(dataLayer),
		UserQuotaRepo: NewUserQuotaRepo(dataLayer),
	}
}
