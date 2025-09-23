package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/zamyatin-zkex/volumer/internal/entity"
)

type State struct {
	kafkaClient sarama.Client
	producer    sarama.SyncProducer
	topic       string
}

func NewState(kafkaClient sarama.Client, prod sarama.SyncProducer, topic string) *State {
	return &State{
		kafkaClient: kafkaClient,
		producer:    prod,
		topic:       topic,
	}
}

func (r *State) LastState(ctx context.Context) (entity.State, error) {
	// Let’s assume for simplicity that we have one partition
	state := entity.State{
		Tokens: make(map[string]entity.Token),
	}

	next, err := r.kafkaClient.GetOffset(r.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return state, fmt.Errorf("get offset: %s", err)
	}
	if next <= 0 {
		// empty topic
		return state, nil
	}

	cons, err := sarama.NewConsumerFromClient(r.kafkaClient)
	if err != nil {
		return state, fmt.Errorf("new consumer: %s", err)
	}
	defer cons.Close()

	cp, err := cons.ConsumePartition(r.topic, 0, sarama.OffsetOldest)
	if err != nil {
		return state, fmt.Errorf("consume partition: %s", err)
	}
	defer cp.Close()

	last := next - 1
	for {
		select {
		case <-ctx.Done():
			return state, ctx.Err()
		case msg := <-cp.Messages():
			token := entity.Token{}
			err := json.Unmarshal(msg.Value, &token)
			if err != nil {
				return state, fmt.Errorf("unmarshal: %s", err)
			}
			state.Tokens[token.Name] = token
			state.Offset = token.Offset

			if msg.Offset == last {
				return state, nil
			}
		}
	}
}

func (s *State) Store(ctx context.Context, state entity.State) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(state.Tokens))
	for name, token := range state.Tokens {
		token.Offset = state.Offset
		payload, err := json.Marshal(token)
		if err != nil {
			return fmt.Errorf("marshal: %s", err)
		}

		msgs = append(msgs, &sarama.ProducerMessage{
			Topic: s.topic,
			Key:   sarama.StringEncoder(name),
			Value: sarama.ByteEncoder(payload),
		})
	}

	return s.producer.SendMessages(msgs)
}
