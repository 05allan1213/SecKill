package data

import (
	"context"
	"encoding/json"
	"github.com/BitofferHub/seckill/internal/log"
)

type SecKillMsgRepo struct {
	data *Data
}

func NewSecKillMsgRepo(data *Data) *SecKillMsgRepo {
	return &SecKillMsgRepo{
		data: data,
	}
}

func (r *SecKillMsgRepo) SendSecKillMsg(ctx context.Context, data *Data, msg *SeckillMessage) error {
	producer := data.GetMQProducer()
	msgJson, err := json.Marshal(msg)
	if err != nil {
		log.ErrorContextf(ctx, "json marshal err %s", err.Error())
		return err
	}
	return producer.SendMessage(ctx, msgJson)
}

func (r *SecKillMsgRepo) UnmarshalSecKillMsg(ctx context.Context, dt *Data, msg []byte) (*SeckillMessage, error) {
	var skMsg = new(SeckillMessage)
	err := json.Unmarshal(msg, skMsg)
	if err != nil {
		return skMsg, err
	}
	return skMsg, err
}
