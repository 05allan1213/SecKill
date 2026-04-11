package model

import (
	"context"
	"encoding/json"

	"github.com/BitofferHub/seckill/internal/log"
)

type SeckillMessage struct {
	TraceID string
	Goods   *Goods
	SecNum  string
	UserID  int64
	Num     int
}

type SecKillMsgModel struct {
	store *Store
}

func NewSecKillMsgModel(store *Store) *SecKillMsgModel {
	return &SecKillMsgModel{store: store}
}

func (m *SecKillMsgModel) WithStore(store *Store) *SecKillMsgModel {
	if store == nil {
		store = m.store
	}
	return &SecKillMsgModel{store: store}
}

func (m *SecKillMsgModel) SendSecKillMsg(ctx context.Context, msg *SeckillMessage) error {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		log.ErrorContextf(ctx, "json marshal err %s", err.Error())
		return err
	}
	return m.store.GetMQProducer().SendMessage(ctx, msgJSON)
}

func (m *SecKillMsgModel) UnmarshalSecKillMsg(ctx context.Context, msg []byte) (*SeckillMessage, error) {
	skMsg := new(SeckillMessage)
	err := json.Unmarshal(msg, skMsg)
	if err != nil {
		return skMsg, err
	}
	return skMsg, nil
}
