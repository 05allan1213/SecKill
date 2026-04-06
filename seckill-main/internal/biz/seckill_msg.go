package biz

import "context"

type SeckillMessage struct {
	TraceID string
	Goods   *Goods
	SecNum  string
	UserID  int64
	Num     int
}

// SecKillMsgRepo is a Greater repo.
type SecKillMsgRepo interface {
	SendSecKillMsg(context.Context, *Data, *SeckillMessage) error
	UnmarshalSecKillMsg(context.Context, *Data, []byte) (*SeckillMessage, error)
}

// SecKillMsgUsecase is a SecKillMsg usecase.
type SecKillMsgUsecase struct {
	repo SecKillMsgRepo
}

// NewSecKillMsgUsecase new a SecKillMsg usecase.
func NewSecKillMsgUsecase(repo SecKillMsgRepo) *SecKillMsgUsecase {
	return &SecKillMsgUsecase{repo: repo}
}

func (uc *SecKillMsgUsecase) SendSecKillMsg(ctx context.Context, data *Data, msg *SeckillMessage) error {
	return uc.repo.SendSecKillMsg(ctx, data, msg)
}

func (uc *SecKillMsgUsecase) UnmarshalSecKillMsg(ctx context.Context, data *Data, msg []byte) (*SeckillMessage, error) {
	return uc.repo.UnmarshalSecKillMsg(ctx, data, msg)
}
