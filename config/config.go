package config

import "github.com/IBM/sarama"

type Config struct {
	Kafka Kafka
	Web   Web
}

// Build
// todo build somehow
func Build() *Config {
	return &Config{
		Kafka{
			TradeTopic: "trades",
			TradeGroup: "volumer",
			RollsTopic: "rolls",
			Brokers:    []string{"127.0.0.1:9092"},
		},
		Web{
			Addr: "127.0.0.1:4242",
		},
	}
}

type Kafka struct {
	TradeTopic string
	TradeGroup string
	RollsTopic string
	Brokers    []string
}

func (k Kafka) SaramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	return cfg
}

type Web struct {
	Addr string
}
