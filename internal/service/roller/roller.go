package roller

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/zamyatin-zkex/volumer/internal/entity"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"
	"github.com/zamyatin-zkex/volumer/pkg/ringbuf"
	"maps"
	"sync"
	"time"
)

type Roller struct {
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

func NewRoller(rest Restorer, eBus *ebus.EBus) *Roller {
	return &Roller{
		tokens:   make(map[string]*Token),
		restored: make(chan struct{}),
		eBus:     eBus,
		restorer: rest,
	}
}

func (r *Roller) AddToken(token *Token) *Roller {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.tokens[token.Name] = token
	return r
}

func (r *Roller) HandleTrade(ctx context.Context, trade event.TradeReceived) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.restored:
	}

	r.mx.RLock()
	defer r.mx.RUnlock()

	// assume pair represents a token
	token, ok := r.tokens[trade.Pair]
	if !ok {
		return fmt.Errorf("token %s not found", trade.Pair)
	}

	if trade.Offset <= token.Offset {
		// skip restored trades
		_ = r.eBus.Emit(ctx, event.TradeSkipped{Offset: trade.Offset})
		return nil
	}

	err := token.Inc(trade)
	if err != nil {
		return fmt.Errorf("failed to inc token %s: %w", trade.Pair, err)
	}

	return nil
}

func (r *Roller) Run(ctx context.Context) error {
	state, err := r.restorer.LastState(ctx)
	if err != nil {
		return fmt.Errorf("restorer state: %w", err)
	}

	err = r.Restore(state)
	if err != nil {
		return fmt.Errorf("restorer restore: %w", err)
	}

	_ = r.eBus.Emit(ctx, event.StateRestored{Offset: state.Offset})

	tradeTicker := time.NewTicker(time.Millisecond * 700)
	defer tradeTicker.Stop()

	stateTicker := time.NewTicker(time.Second * 5)
	defer stateTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tradeTicker.C:
			r.mx.RLock()
			for _, token := range r.tokens {
				// ensure there are ontime buckets without trades
				_ = token.Inc(event.TradeReceived{
					Trade: entity.Trade{
						Volume: decimal.Zero,
						Time:   time.Now(),
					},
					Offset: 0,
				})
			}
			r.mx.RUnlock()
		case <-stateTicker.C:
			currentState := r.State()
			err = r.restorer.Store(ctx, currentState)
			if err != nil {
				return fmt.Errorf("restorer store: %w", err)
			}

			//os.Exit(33)
			err = r.eBus.Emit(ctx, event.StateSaved{
				Offset: currentState.Offset,
			})
			if err != nil {
				return fmt.Errorf("ebus emit: %w", err)
			}
		}
	}
}

func (r *Roller) Stats() map[string]map[string]decimal.Decimal {
	r.mx.RLock()
	defer r.mx.RUnlock()

	stats := make(map[string]map[string]decimal.Decimal)

	for _, token := range r.tokens {
		stats[token.Name] = token.Stats()
	}

	return stats
}

func (r *Roller) State() entity.State {
	r.mx.RLock()
	defer r.mx.RUnlock()

	state := entity.State{
		Tokens: map[string]entity.Token{},
	}

	for _, token := range r.tokens {
		periods := make(map[string]time.Duration)
		maps.Copy(periods, token.Periods)

		sums := make(map[string]decimal.Decimal)
		maps.Copy(sums, token.RollSums)

		buckets := ringbuf.Ring[entity.Bucket]{}
		buckets.Head = token.Buckets.Head
		buckets.Data = make([]entity.Bucket, token.Buckets.Len())
		for i := 0; i < token.Buckets.Len(); i++ {
			buckets.Data[i] = entity.Bucket{
				StartedAt: token.Buckets.Data[i].StartedAt,
				Volume:    token.Buckets.Data[i].Volume,
			}
		}

		state.Tokens[token.Name] = entity.Token{
			Name:     token.Name,
			Periods:  periods,
			Buckets:  &buckets,
			RollSums: sums,
		}

		if token.Offset > state.Offset {
			state.Offset = token.Offset
		}
	}

	return state
}

func (r *Roller) Restore(state entity.State) error {
	defer close(r.restored)

	r.mx.Lock()
	defer r.mx.Unlock()

	for _, token := range state.Tokens {
		_, ok := r.tokens[token.Name]
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

		r.tokens[token.Name] = &Token{
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
