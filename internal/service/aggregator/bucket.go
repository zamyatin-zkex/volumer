package aggregator

import (
	"time"

	"github.com/shopspring/decimal"
)

type Bucket struct {
	StartedAt time.Time
	Volume    decimal.Decimal
}

func (b Bucket) add(volume decimal.Decimal) Bucket {
	b.Volume = b.Volume.Add(volume)
	return b
}
