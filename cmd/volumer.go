package main

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/zamyatin-zkex/volumer/config"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"github.com/zamyatin-zkex/volumer/internal/repository"
	"github.com/zamyatin-zkex/volumer/internal/service/aggregator"
	"github.com/zamyatin-zkex/volumer/internal/service/consumer"
	"github.com/zamyatin-zkex/volumer/internal/service/faketrader"
	"github.com/zamyatin-zkex/volumer/internal/service/watcher"
	"github.com/zamyatin-zkex/volumer/internal/service/web"
	"github.com/zamyatin-zkex/volumer/pkg/ebus"
	"github.com/zamyatin-zkex/volumer/pkg/utils"
	"log"
	"time"

	"github.com/zamyatin-zkex/volumer/internal/service/interrupter"
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
	aggregator := aggregator.NewAggregator(stateRepo, eBus).
		AddToken(aggregator.NewToken("ETH", aggregator.Periods{"3s": time.Second * 3, "10s": time.Second * 10})).
		AddToken(aggregator.NewToken("BTC", aggregator.Periods{"10s": time.Second * 10, "30s": time.Second * 30}))
	watch := watcher.NewWatcher(eBus).
		EmitEvery(time.Second, event.StatsUpdated{}, func(ctx context.Context) (any, error) {
			return event.StatsUpdated{Tokens: aggregator.Stats()}, nil
		})

	eBus.
		Subscribe(event.StateSaved{}, watcher.LogAny).
		Subscribe(event.StateRestored{}, watcher.LogAny).
		Subscribe(event.TradeSkipped{}, watcher.LogAny).
		Subscribe(event.StateSaved{}, ebus.Typed(consumer.Commit)).
		Subscribe(event.TradeReceived{}, ebus.Typed(aggregator.HandleTrade)).
		Subscribe(event.StatsUpdated{}, ebus.Typed(web.UpdateStats))

	err := app.NewApp().
		WithService(fakeTrader).
		WithService(consumer).
		WithService(aggregator).
		WithService(watch).
		WithService(web).
		WithService(interrupter.Interrupter{}).
		Run(context.Background())

	log.Fatal(err)
}
