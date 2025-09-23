package event

import (
	"github.com/shopspring/decimal"
)

type StatsUpdated struct {
	Tokens map[string]map[string]decimal.Decimal
}
