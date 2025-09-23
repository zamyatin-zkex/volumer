package faketrader

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/zamyatin-zkex/volumer/internal/entity"
	"math/rand"
	"time"
)

type TradeStore interface {
	Store(ctx context.Context, trade entity.Trade) error
}

type Trader struct {
	tokens []string
	repo   TradeStore
}

func NewTrader(repo TradeStore, tokens ...string) *Trader {
	return &Trader{
		repo:   repo,
		tokens: tokens,
	}
}

func (t *Trader) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, token := range t.tokens {
				val := rand.Intn(200)
				trade := entity.Trade{
					Pair:   token,
					Volume: decimal.NewFromInt(int64(val)),
					Time:   time.Now(),
				}

				err := t.repo.Store(ctx, trade)
				if err != nil {
					return err
				}
			}
		}
	}
}
