package consumer

import (
	"context"
	"fmt"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	handler       Handler
	eBus          *ebus.EBus
}

func NewConsumer(client sarama.Client, topic string, group string, eBus *ebus.EBus) (*Consumer, error) {
	cons, err := sarama.NewConsumerGroupFromClient(group, client)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: cons,
		handler: Handler{
			commits: make(chan int64),
			topic:   topic,
			eBus:    eBus,
		},
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, 1)

	go func() {
		for {
			if err := c.consumerGroup.Consume(ctx, c.handler.topics(), c.handler); err != nil {
				errs <- err
				return
			}

			if ctx.Err() != nil {
				errs <- ctx.Err()
				return
			}
		}
	}()

	select {
	case err := <-errs:
		return fmt.Errorf("consumer error: %w", err)
	case err := <-c.consumerGroup.Errors():
		return fmt.Errorf("consumerGroup error: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("consumer: %w", ctx.Err())
	}
}

func (c *Consumer) Commit(ctx context.Context, saved event.StateSaved) error {
	//return nil
	return c.handler.commit(ctx, saved.Offset)
}
