package aggregator

import (
	"github.com/zamyatin-zkex/volumer/internal/entity"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"maps"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zamyatin-zkex/volumer/pkg/ringbuf"
)

type Token struct {
	Name      string
	Periods   Periods
	Buckets   *ringbuf.Ring[Bucket]
	RollSums  map[string]decimal.Decimal
	mx        sync.RWMutex
	Offset    int64
	Partition int64
}

func NewToken(name string, periods Periods) *Token {
	return &Token{
		Name:     name,
		Periods:  periods,
		Buckets:  ringbuf.New[Bucket](periods.max().buckets()).PushFront(Bucket{StartedAt: time.Now()}),
		RollSums: make(map[string]decimal.Decimal),
	}
}

func (t *Token) Inc(trade event.TradeReceived) error {
	t.mx.Lock()
	defer t.mx.Unlock()

	amount := trade.Volume
	ts := trade.Time

	lastBucket := t.Buckets.GetN(0)

	// late arrived event
	if ts.Before(lastBucket.StartedAt) {
		// todo
		// find the bucket and update periods vals
		// if bucked has been deleted (time < max bucket) send to DLQ
		return nil
	}

	// fall into the last bucket
	if ts.Before(lastBucket.StartedAt.Add(time.Second)) {
		lastBucket = lastBucket.add(amount)
		t.Buckets.SetN(0, lastBucket)

		for name := range t.Periods {
			t.RollSums[name] = t.RollSums[name].Add(amount)
		}

		return nil
	}

	// create new bucket
	for name, dur := range t.Periods {
		oldBucket := t.Buckets.GetN(Period(dur).buckets() - 1)
		t.RollSums[name] = t.RollSums[name].Sub(oldBucket.Volume).Add(amount)
	}

	t.Buckets.PushFront(Bucket{
		StartedAt: lastBucket.StartedAt.Add(time.Second),
		Volume:    amount,
	})

	if trade.Offset > 0 {
		t.Offset = trade.Offset
	}

	return nil
}

func (t *Token) stats() map[string]decimal.Decimal {
	t.mx.RLock()
	defer t.mx.RUnlock()

	stats := make(map[string]decimal.Decimal)
	maps.Copy(stats, t.RollSums)

	return stats
}

func (t *Token) state() entity.Token {
	t.mx.RLock()
	defer t.mx.RUnlock()

	periods := make(map[string]time.Duration)
	maps.Copy(periods, t.Periods)

	sums := make(map[string]decimal.Decimal)
	maps.Copy(sums, t.RollSums)

	buckets := ringbuf.Ring[entity.Bucket]{}
	buckets.Head = t.Buckets.Head
	buckets.Data = make([]entity.Bucket, t.Buckets.Len())
	for i := 0; i < t.Buckets.Len(); i++ {
		buckets.Data[i] = entity.Bucket{
			StartedAt: t.Buckets.Data[i].StartedAt,
			Volume:    t.Buckets.Data[i].Volume,
		}
	}

	return entity.Token{
		Name:     t.Name,
		Periods:  periods,
		Buckets:  &buckets,
		RollSums: sums,
		Offset:   t.Offset,
	}
}
