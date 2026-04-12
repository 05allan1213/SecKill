package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/BitofferHub/seckill/internal/log"
)

type SeckillEnvelope struct {
	Payload      *SeckillMessage `json:"payload"`
	Attempt      int             `json:"attempt"`
	FirstSentAt  int64           `json:"first_sent_at"`
	LastError    string          `json:"last_error,omitempty"`
	SourceTopic  string          `json:"source_topic,omitempty"`
}

type SecKillMsgRepo struct {
	data *Data
}

func NewSecKillMsgRepo(data *Data) *SecKillMsgRepo {
	return &SecKillMsgRepo{
		data: data,
	}
}

func NewSeckillEnvelope(msg *SeckillMessage) *SeckillEnvelope {
	return &SeckillEnvelope{
		Payload:     msg,
		Attempt:     0,
		FirstSentAt: time.Now().UnixMilli(),
	}
}

func (e *SeckillEnvelope) IncrementAttempt(lastError string) *SeckillEnvelope {
	return &SeckillEnvelope{
		Payload:     e.Payload,
		Attempt:     e.Attempt + 1,
		FirstSentAt: e.FirstSentAt,
		LastError:   lastError,
		SourceTopic: e.SourceTopic,
	}
}

func (r *SecKillMsgRepo) SendSecKillMsg(ctx context.Context, data *Data, msg *SeckillMessage) error {
	producer := data.GetMQProducer()
	envelope := NewSeckillEnvelope(msg)
	msgJson, err := json.Marshal(envelope)
	if err != nil {
		log.Error(ctx, "marshal seckill envelope failed",
			log.Field(log.FieldAction, "mq.marshal"),
			log.Field(log.FieldError, err.Error()),
		)
		return err
	}
	return producer.SendMessage(ctx, msgJson)
}

func (r *SecKillMsgRepo) UnmarshalSecKillMsg(ctx context.Context, dt *Data, msg []byte) (*SeckillMessage, error) {
	var envelope SeckillEnvelope
	if err := json.Unmarshal(msg, &envelope); err == nil && envelope.Payload != nil {
		return envelope.Payload, nil
	}
	
	var legacyMsg SeckillMessage
	if err := json.Unmarshal(msg, &legacyMsg); err != nil {
		return nil, err
	}
	return &legacyMsg, nil
}

func (r *SecKillMsgRepo) UnmarshalEnvelope(ctx context.Context, msg []byte) (*SeckillEnvelope, error) {
	var envelope SeckillEnvelope
	if err := json.Unmarshal(msg, &envelope); err != nil {
		var legacyMsg SeckillMessage
		if legacyErr := json.Unmarshal(msg, &legacyMsg); legacyErr != nil {
			return nil, err
		}
		return &SeckillEnvelope{
			Payload:     &legacyMsg,
			Attempt:     0,
			FirstSentAt: time.Now().UnixMilli(),
		}, nil
	}
	if envelope.Payload == nil {
		return nil, nil
	}
	return &envelope, nil
}

func (r *SecKillMsgRepo) SendToRetry(ctx context.Context, data *Data, envelope *SeckillEnvelope, lastError string) error {
	retryProducer := data.GetRetryProducer()
	if retryProducer == nil {
		log.Error(ctx, "retry producer not configured",
			log.Field(log.FieldAction, "mq.retry.send"),
		)
		return errors.New("retry producer not configured")
	}

	retryEnvelope := envelope.IncrementAttempt(lastError)
	retryEnvelope.SourceTopic = data.conf.Kafka.Consumer.Topic

	envelopeJson, err := json.Marshal(retryEnvelope)
	if err != nil {
		log.Error(ctx, "marshal retry envelope failed",
			log.Field(log.FieldAction, "mq.retry.marshal"),
			log.Field(log.FieldError, err.Error()),
		)
		return err
	}

	return retryProducer.SendMessage(ctx, envelopeJson)
}

func (r *SecKillMsgRepo) SendToDLQ(ctx context.Context, data *Data, envelope *SeckillEnvelope, lastError string) error {
	dlqProducer := data.GetDLQProducer()
	if dlqProducer == nil {
		log.Error(ctx, "dlq producer not configured",
			log.Field(log.FieldAction, "mq.dlq.send"),
		)
		return errors.New("dlq producer not configured")
	}

	dlqEnvelope := envelope.IncrementAttempt(lastError)
	dlqEnvelope.SourceTopic = data.conf.Kafka.Consumer.Topic

	envelopeJson, err := json.Marshal(dlqEnvelope)
	if err != nil {
		log.Error(ctx, "marshal dlq envelope failed",
			log.Field(log.FieldAction, "mq.dlq.marshal"),
			log.Field(log.FieldError, err.Error()),
		)
		return err
	}

	return dlqProducer.SendMessage(ctx, envelopeJson)
}
