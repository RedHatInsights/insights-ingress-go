package queue

import (
	"context"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer consumes in and produces to the topic in config
// Each message is sent to the writer via a goroutine so that the internal batch
// buffer has an opportunity to fill.
func Producer(in chan []byte, config *ProducerConfig) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		Balancer: &kafka.Hash{},
		Async:    config.Async,
	})

	defer w.Close()

	for v := range in {
		go func(v []byte) {
			err := w.WriteMessages(context.Background(),
				kafka.Message{
					Key:   nil,
					Value: v,
				},
			)
			if err != nil {
				l.Log.Error("error while writing, putting message back into the channel", zap.Error(err))
				in <- v
			}
		}(v)
	}
}
