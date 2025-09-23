package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/zamyatin-zkex/volumer/internal/entity"
)

type Trade struct {
	producer sarama.SyncProducer
	topic    string
}

func NewTrade(producer sarama.SyncProducer, topic string) *Trade {
	return &Trade{producer: producer, topic: topic}
}

func (t Trade) Store(ctx context.Context, trade entity.Trade) error {
	js, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("json marshal trade: %w", err)
	}

	_, _, err = t.producer.SendMessage(&sarama.ProducerMessage{
		Topic: t.topic,
		Key:   sarama.StringEncoder(trade.Pair),
		Value: sarama.ByteEncoder(js),
	})
	if err != nil {
		return fmt.Errorf("send trade to kafka: %w", err)
	}

	return nil
}
