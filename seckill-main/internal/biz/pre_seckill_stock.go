package biz

import "context"

// SecKillStockRepo is a Greater repo.
type PreSecKillStockRepo interface {
	PreDescStock(context.Context, *Data, int64, int64, int32, string, *PreSecKillRecord) (string, error)
	SetSuccessInPreSecKill(context.Context, *Data,
		int64, int64, string, *PreSecKillRecord) (string, error)
	GetSecKillInfo(context.Context, *Data, string) (*PreSecKillRecord, error)
	//	Try(ctx context.Context, data *Data, a int) error
	//DescStock(ctx context.Context, data *Data, goodsID int64, num int64) (int64, error)
}

// SecKillStockUsecase is a SecKillStock usecase.
type PreSecKillStockUsecase struct {
	repo PreSecKillStockRepo
}

// NewSecKillStockUsecase new a SecKillStock usecase.
func NewPreSecKillStockUsecase(repo PreSecKillStockRepo) *PreSecKillStockUsecase {
	return &PreSecKillStockUsecase{repo: repo}
}

func (uc *PreSecKillStockUsecase) PreDescStock(ctx context.Context, data *Data, userID int64, goodsID int64,
	num int32, secNum string, record *PreSecKillRecord) (string, error) {
	return uc.repo.PreDescStock(ctx, data, userID, goodsID, num, secNum, record)
}

func (uc *PreSecKillStockUsecase) GetSecKillInfo(ctx context.Context, data *Data, secNum string) (*PreSecKillRecord, error) {
	return uc.repo.GetSecKillInfo(ctx, data, secNum)
}

func (uc *PreSecKillStockUsecase) SetSuccessInPreSecKill(ctx context.Context, data *Data,
	userID int64, goodsID int64, secNum string, record *PreSecKillRecord) (string, error) {
	return uc.repo.SetSuccessInPreSecKill(ctx, data, userID, goodsID, secNum, record)
}
