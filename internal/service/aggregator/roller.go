package aggregator

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/zamyatin-zkex/volumer/internal/entity"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"
	"github.com/zamyatin-zkex/volumer/pkg/ringbuf"
	"sync"
	"time"
)

type Aggregator struct {
	mx sync.RWMutex

	tokens map[string]*Token

	restorer Restorer
	restored chan struct{}

	eBus *ebus.EBus
}

type Restorer interface {
	LastState(context.Context) (entity.State, error)
	Store(context.Context, entity.State) error
}

func NewAggregator(rest Restorer, eBus *ebus.EBus) *Aggregator {
	return &Aggregator{
		tokens:   make(map[string]*Token),
		restored: make(chan struct{}),
		eBus:     eBus,
		restorer: rest,
	}
}

func (a *Aggregator) AddToken(token *Token) *Aggregator {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.tokens[token.Name] = token
	return a
}

func (a *Aggregator) HandleTrade(ctx context.Context, trade event.TradeReceived) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-a.restored:
	}

	a.mx.RLock()
	defer a.mx.RUnlock()

	// assume pair represents a token
	token, ok := a.tokens[trade.Pair]
	if !ok {
		return fmt.Errorf("token %s not found", trade.Pair)
	}

	if trade.Offset <= token.Offset {
		// skip restored trades
		_ = a.eBus.Emit(ctx, event.TradeSkipped{Offset: trade.Offset})
		return nil
	}

	err := token.Inc(trade)
	if err != nil {
		return fmt.Errorf("failed to inc token %s: %w", trade.Pair, err)
	}

	return nil
}

func (a *Aggregator) Run(ctx context.Context) error {
	state, err := a.restorer.LastState(ctx)
	if err != nil {
		return fmt.Errorf("restorer state: %w", err)
	}

	err = a.restore(state)
	if err != nil {
		return fmt.Errorf("restorer restore: %w", err)
	}

	_ = a.eBus.Emit(ctx, event.StateRestored{Offset: state.Offset})

	tradeTicker := time.NewTicker(time.Millisecond * 700)
	defer tradeTicker.Stop()

	stateTicker := time.NewTicker(time.Second * 5)
	defer stateTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tradeTicker.C:
			a.mx.RLock()
			for _, token := range a.tokens {
				// ensure there are ontime buckets without trades
				_ = token.Inc(event.TradeReceived{
					Trade: entity.Trade{
						Volume: decimal.Zero,
						Time:   time.Now(),
					},
					Offset: 0,
				})
			}
			a.mx.RUnlock()
		case <-stateTicker.C:
			currentState := a.state()
			err = a.restorer.Store(ctx, currentState)
			if err != nil {
				return fmt.Errorf("restorer store: %w", err)
			}

			//os.Exit(33)
			err = a.eBus.Emit(ctx, event.StateSaved{
				Offset: currentState.Offset,
			})
			if err != nil {
				return fmt.Errorf("ebus emit: %w", err)
			}
		}
	}
}

func (a *Aggregator) Stats() map[string]map[string]decimal.Decimal {
	a.mx.RLock()
	defer a.mx.RUnlock()

	stats := make(map[string]map[string]decimal.Decimal)

	for _, token := range a.tokens {
		stats[token.Name] = token.stats()
	}

	return stats
}

func (a *Aggregator) state() entity.State {
	a.mx.RLock()
	defer a.mx.RUnlock()

	state := entity.State{
		Tokens: map[string]entity.Token{},
	}

	for _, token := range a.tokens {
		tokenState := token.state()
		state.Tokens[token.Name] = tokenState

		if tokenState.Offset > state.Offset {
			state.Offset = tokenState.Offset
		}
	}

	return state
}

func (a *Aggregator) restore(state entity.State) error {
	defer close(a.restored)

	a.mx.Lock()
	defer a.mx.Unlock()

	for _, token := range state.Tokens {
		_, ok := a.tokens[token.Name]
		if ok {
			// todo
			// some merge logic
			// continue
		}

		buckets := ringbuf.Ring[Bucket]{}
		buckets.Head = token.Buckets.Head
		buckets.Data = make([]Bucket, token.Buckets.Len())
		for i := 0; i < token.Buckets.Len(); i++ {
			buckets.Data[i] = Bucket{
				StartedAt: token.Buckets.Data[i].StartedAt,
				Volume:    token.Buckets.Data[i].Volume,
			}
		}

		a.tokens[token.Name] = &Token{
			Name:     token.Name,
			Periods:  token.Periods,
			Buckets:  &buckets,
			RollSums: token.RollSums,
			mx:       sync.RWMutex{},
			Offset:   token.Offset,
		}
	}
	return nil
}
