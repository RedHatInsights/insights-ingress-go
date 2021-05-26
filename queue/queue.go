package queue

import (
	"context"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"

	p "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	messagesPublished = pa.NewCounterVec(p.CounterOpts{
		Name: "ingress_kafka_produced",
		Help: "Number of messages produced to kafka",
	}, []string{"topic"})
	messagePublishElapsed = pa.NewHistogramVec(p.HistogramOpts{
		Name: "ingress_publish_seconds",
		Help: "Number of seconds spent writing kafka messages",
	}, []string{"topic"})
	publishFailures = pa.NewCounterVec(p.CounterOpts{
		Name: "ingress_kafka_produce_failures",
		Help: "Number of times a message was failed to be produced",
	}, []string{"topic"})
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
			start := time.Now()
			err := w.WriteMessages(context.Background(),
				kafka.Message{
					Key:   nil,
					Value: v,
				},
			)
			messagePublishElapsed.With(p.Labels{"topic": config.Topic}).Observe(time.Since(start).Seconds())
			if err != nil {
				l.Log.WithFields(logrus.Fields{"error": err}).Error("error while writing, putting message back into the channel")
				in <- v
				publishFailures.With(p.Labels{"topic": config.Topic}).Inc()
				return
			}

			messagesPublished.With(p.Labels{"topic": config.Topic}).Inc()
		}(v)
	}
}
