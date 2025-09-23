package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"os"
	"os/signal"
)

func main() {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	consumer, err := sarama.NewConsumer([]string{"127.0.0.1:9092"}, cfg)
	if err != nil {
		panic(err)
	}

	defer consumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	partitionConsumer, err := consumer.ConsumePartition("trades", 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Println(msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
		case <-signals:
			return
		}
	}
}
