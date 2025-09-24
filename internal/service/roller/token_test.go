package roller

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/zamyatin-zkex/volumer/internal/entity"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"testing"
	"time"
)

func TestToken(t *testing.T) {
	token := NewToken("mycoin", Periods{"1s": time.Second, "2s": time.Second * 2})

	ten := decimal.NewFromInt(10)

	trade := event.TradeReceived{
		Trade: entity.Trade{
			Volume: ten,
			Time:   time.Now(),
		},
	}
	err := token.Inc(trade)
	assert.NoError(t, err)
	assert.Equal(t, ten, token.RollSums["1s"])
	assert.Equal(t, ten, token.RollSums["2s"])

	trade.Time = trade.Time.Add(time.Second)
	err = token.Inc(trade)
	assert.NoError(t, err)
	assert.Equal(t, ten, token.RollSums["1s"])
	assert.Equal(t, ten.Add(ten), token.RollSums["2s"])

	trade.Time = trade.Time.Add(time.Second)
	err = token.Inc(trade)
	assert.NoError(t, err)
	assert.Equal(t, ten, token.RollSums["1s"])
	assert.Equal(t, ten.Add(ten), token.RollSums["2s"])
}
