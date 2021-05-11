package queue

import (
	"time"

	l "github.com/redhatinsights/insights-ingress-go/pkg/logger"
	"github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

	prom "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	messagesPublished = pa.NewCounterVec(prom.CounterOpts{
		Name: "ingress_kafka_produced",
		Help: "Number of messages produced to kafka",
	}, []string{"topic"})
	messagePublishElapsed = pa.NewHistogramVec(prom.HistogramOpts{
		Name: "ingress_publish_seconds",
		Help: "Number of seconds spent writing kafka messages",
	}, []string{"topic"})
	publishFailures = pa.NewCounterVec(prom.CounterOpts{
		Name: "ingress_kafka_produce_failures",
		Help: "Number of times a message was failed to be produced",
	}, []string{"topic"})
)

// Producer consumes in and produces to the topic in config
// Each message is sent to the writer via a goroutine so that the internal batch
// buffer has an opportunity to fill.
func Producer(in chan []byte, config *ProducerConfig) {
	configMap := kafka.ConfigMap{
		"bootstrap.servers": config.Brokers[0],
	}

	if config.CA != "" {
		configMap["ssl.ca.location"] = config.CA
	}

	if config.Username != "" {
		configMap["sasl.username"] = config.Username
		configMap["sasl.password"] = config.Password
		configMap["sasl.mechanism"] = "SCRAM-SHA-512"
		configMap["security.protocol"] = "SASL_SSL"
	}

	p, err := kafka.NewProducer(&configMap)

	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("Error creating kafka producer")
		// TODO: Somehow indicate to caller that this failed
		return
	}

	defer p.Close()

	for v := range in {
		go func(v []byte) {
			start := time.Now()
			err := p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &config.Topic,
					Partition: kafka.PartitionAny,
				},
				Value: v,
			}, nil)
			messagePublishElapsed.With(prom.Labels{"topic": config.Topic}).Observe(time.Since(start).Seconds())
			if err != nil {
				l.Log.WithFields(logrus.Fields{"error": err}).Error("error while writing, putting message back into the channel")
				in <- v
				publishFailures.With(prom.Labels{"topic": config.Topic}).Inc()
				return
			}

			messagesPublished.With(prom.Labels{"topic": config.Topic}).Inc()
		}(v)
	}
}
