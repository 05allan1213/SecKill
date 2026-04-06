package data

import (
	"context"
	"encoding/json"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
)

type secKillMsgRepo struct {
	data *Data
}

// NewSecKillStockRepo
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param data
//	@return biz.SecKillStockRepo
func NewSecKillMsgRepo(data *Data) biz.SecKillMsgRepo {
	return &secKillMsgRepo{
		data: data,
	}
}

func (r *secKillMsgRepo) SendSecKillMsg(ctx context.Context, data *biz.Data, msg *biz.SeckillMessage) error {
	producer := data.GetMQProducer()
	msgJson, err := json.Marshal(msg)
	if err != nil {
		log.ErrorContextf(ctx, "json marshal err %s", err.Error())
		return err
	}
	return producer.SendMessage(msgJson)
}

func (r *secKillMsgRepo) UnmarshalSecKillMsg(ctx context.Context,
	dt *biz.Data, msg []byte) (*biz.SeckillMessage, error) {
	var skMsg = new(biz.SeckillMessage)
	err := json.Unmarshal(msg, skMsg)
	if err != nil {
		return skMsg, err
	}
	return skMsg, err
}
