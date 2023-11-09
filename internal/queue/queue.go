package queue

import (
	"strings"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
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

// ProducerConfig configures a producer
type ProducerConfig struct {
	Topic                 string
	Brokers               []string
	Async                 bool
	Username              string
	Password              string
	CA                    string
	KafkaSecurityProtocol string
	SASLMechanism         string
	KafkaDeliveryReports  bool
	Debug                 bool
}

// Producer consumes in and produces to the topic in config
// Each message is sent to the writer via a goroutine so that the internal batch
// buffer has an opportunity to fill.
func Producer(in chan validators.ValidationMessage, config *ProducerConfig) {

	var configMap kafka.ConfigMap

	configMap = kafka.ConfigMap{
		"bootstrap.servers":   strings.Join(config.Brokers, ","),
		"go.delivery.reports": config.KafkaDeliveryReports,
	}

	if config.CA != "" {
		_ = configMap.SetKey("ssl.ca.location", config.CA)
	}

	if config.SASLMechanism != "" {
		_ = configMap.SetKey("security.protocol", config.KafkaSecurityProtocol)
		_ = configMap.SetKey("sasl.mechanism", config.SASLMechanism)
		_ = configMap.SetKey("sasl.username", config.Username)
		_ = configMap.SetKey("sasl.password", config.Password)
	}

	if config.Debug {
		_ = configMap.SetKey("debug", "protocol,broker,topic")
	}

	p, err := kafka.NewProducer(&configMap)

	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("Error creating kafka producer")
		return
	}

	defer p.Close()

	for v := range in {
		go func(v validators.ValidationMessage) {
			delivery_chan := make(chan kafka.Event)
			defer close(delivery_chan)
			producerCount.Inc()
			defer producerCount.Dec()
			start := time.Now()
			kafkaHeaders := make([]kafka.Header, len(v.Headers))
			i := 0
			for key, value := range v.Headers {
				kafkaHeaders[i] = kafka.Header{
					Key:   key,
					Value: []byte(value),
				}
				i++
			}
			p.Produce(&kafka.Message{
				Headers: kafkaHeaders,
				TopicPartition: kafka.TopicPartition{
					Topic:     &config.Topic,
					Partition: kafka.PartitionAny,
				},
				Value: v.Message,
				Key:   v.Key,
			}, delivery_chan)
			messagePublishElapsed.With(prom.Labels{"topic": config.Topic}).Observe(time.Since(start).Seconds())

			e := <-delivery_chan
			m := e.(*kafka.Message)

			if m.TopicPartition.Error != nil {
				l.Log.WithFields(logrus.Fields{"error": m.TopicPartition.Error}).Error("Error publishing to kafka")
				in <- v
				publishFailures.With(prom.Labels{"topic": config.Topic}).Inc()
				return
			} else {
				messagesPublished.With(prom.Labels{"topic": config.Topic}).Inc()
			}
		}(v)
	}
}
