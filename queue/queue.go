package queue

import (
	"time"

	l "github.com/redhatinsights/insights-ingress-go/logger"
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
	producerCount = pa.NewGauge(prom.GaugeOpts{
		Name: "ingress_kafka_producer_go_routine_count",
		Help: "Number of go routines currently publishing to kafka",
	})
)

// Producer consumes in and produces to the topic in config
// Each message is sent to the writer via a goroutine so that the internal batch
// buffer has an opportunity to fill.
func Producer(in chan []byte, config *ProducerConfig) {

	var configMap kafka.ConfigMap

	if config.SASLMechanism != "" {
		configMap = kafka.ConfigMap{
			"bootstrap.servers":   config.Brokers[0],
			"security.protocol":   config.Protocol,
			"sasl.mechanism":      config.SASLMechanism,
			"ssl.ca.location":     config.CA,
			"sasl.username":       config.Username,
			"sasl.password":       config.Password,
			"go.delivery.reports": config.KafkaDeliveryReports,
		}
	} else {
		configMap = kafka.ConfigMap{
			"bootstrap.servers":   config.Brokers[0],
			"go.delivery.reports": config.KafkaDeliveryReports,
		}
	}

	p, err := kafka.NewProducer(&configMap)

	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("Error creating kafka producer")
		return
	}

	defer p.Close()

	// Read the Events() channel for this producer, if an Error exists in the TopicPartition, then put the
	// message back in the queue to be reprocessed. This is to keep messages from getting lost in the event
	// of a brief kafka outage.
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					l.Log.WithFields(logrus.Fields{"error": ev.TopicPartition.Error}).Error("Error publishing to kafka")
					in <- ev.Value
					publishFailures.With(prom.Labels{"topic": config.Topic}).Inc()
				} else {
					messagesPublished.With(prom.Labels{"topic": config.Topic}).Inc()
					l.Log.Info("Message published to kafka")
				}
			}
		}
	}()

	for v := range in {
		go func(v []byte) {
			producerCount.Inc()
			defer producerCount.Dec()
			start := time.Now()
			p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &config.Topic,
					Partition: kafka.PartitionAny,
				},
				Value: v,
			}, nil)
			messagePublishElapsed.With(prom.Labels{"topic": config.Topic}).Observe(time.Since(start).Seconds())
		}(v)
	}
}
