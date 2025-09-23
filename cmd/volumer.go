package main

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/zamyatin-zkex/volumer/config"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"github.com/zamyatin-zkex/volumer/internal/repository"
	"github.com/zamyatin-zkex/volumer/internal/service/consumer"
	"github.com/zamyatin-zkex/volumer/internal/service/faketrader"
	"github.com/zamyatin-zkex/volumer/internal/service/watcher"
	"github.com/zamyatin-zkex/volumer/internal/service/web"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"
	"github.com/zamyatin-zkex/volumer/pkg/utils"
	"log"
	"time"

	"github.com/zamyatin-zkex/volumer/internal/service/interrupter"
	"github.com/zamyatin-zkex/volumer/internal/service/roller"
	"github.com/zamyatin-zkex/volumer/pkg/app"
)

func main() {
	cfg := config.Build()
	eBus := ebus.New()

	kafkaCl := utils.Must(sarama.NewClient(cfg.Kafka.Brokers, cfg.Kafka.SaramaConfig()))
	defer kafkaCl.Close()
	prod := utils.Must(sarama.NewSyncProducerFromClient(kafkaCl))
	defer prod.Close()

	stateRepo := repository.NewState(kafkaCl, prod, cfg.Kafka.RollsTopic)
	tradeRepo := repository.NewTrade(prod, cfg.Kafka.TradeTopic)

	fakeTrader := faketrader.NewTrader(tradeRepo, "ETH", "BTC")
	consumer := utils.Must(consumer.NewConsumer(kafkaCl, cfg.Kafka.TradeTopic, cfg.Kafka.TradeGroup, eBus))
	web := web.New(cfg.Web.Addr)
	roller := roller.NewRoller(stateRepo, eBus).
		AddToken(roller.NewToken("ETH", roller.Periods{"3s": time.Second * 3, "10s": time.Second * 10})).
		AddToken(roller.NewToken("BTC", roller.Periods{"10s": time.Second * 10, "30s": time.Second * 30}))
	watch := watcher.NewWatcher(eBus).
		EmitEvery(time.Second, event.StatsUpdated{}, func(ctx context.Context) (any, error) {
			return event.StatsUpdated{Tokens: roller.Stats()}, nil
		})

	eBus.
		Subscribe(event.StateSaved{}, watcher.LogAny).
		Subscribe(event.StateRestored{}, watcher.LogAny).
		Subscribe(event.TradeSkipped{}, watcher.LogAny).
		Subscribe(event.StateSaved{}, ebus.Typed(consumer.Commit)).
		Subscribe(event.TradeReceived{}, ebus.Typed(roller.HandleTrade)).
		Subscribe(event.StatsUpdated{}, ebus.Typed(web.UpdateStats))

	err := app.NewApp().
		WithService(roller).
		WithService(fakeTrader).
		WithService(watch).
		WithService(consumer).
		WithService(web).
		WithService(interrupter.Interrupter{}).
		Run(context.Background())

	log.Fatal(err)
}
