package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/zamyatin-zkex/volumer/internal/entity"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"
)

var _ sarama.ConsumerGroupHandler = Handler{}

type Handler struct {
	commits chan int64
	topic   string
	eBus    *ebus.EBus
}

func (h Handler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			errs := make(chan error, 1)
			go func() {
				errs <- h.handle(session.Context(), msg)
			}()
			select {
			case err := <-errs:
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					return fmt.Errorf("claim handle: %w", err)
				}
			case <-session.Context().Done():
				return nil
			}

		case <-session.Context().Done():
			return nil

		case offset := <-h.commits:
			session.MarkOffset(h.topic, 0, offset+1, "")
		}

	}
}

func (h Handler) commit(ctx context.Context, offset int64) error {
	select {
	case h.commits <- offset:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h Handler) topics() []string {
	return []string{h.topic}
}

func (h Handler) handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	trade := entity.Trade{}
	if err := json.Unmarshal(message.Value, &trade); err != nil {
		return fmt.Errorf("unmarshal trade: %w", err)
	}

	return h.eBus.Emit(ctx, event.TradeReceived{
		Trade:  trade,
		Offset: message.Offset,
	})
}
