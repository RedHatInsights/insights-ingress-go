package queue

import (
	"context"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

//Producer consumes in and produces to the topic in config
func Producer(in chan []byte, config *ProducerConfig) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		Balancer: &kafka.Hash{},
	})

	defer w.Close()

	for {
		v := <-in
		err := w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   nil,
				Value: v,
			},
		)
		if err != nil {
			l.Log.Error("error while writing", zap.Error(err))
		}
	}
}

// Consumer consumes a topic and puts the messages into out
func Consumer(out chan []byte, config *ConsumerConfig) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		GroupID: config.GroupID,
		Topic:   config.Topic,
	})

	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			l.Log.Error("Error while reading message", zap.Error(err))
		} else {
			out <- m.Value
		}
	}
}
