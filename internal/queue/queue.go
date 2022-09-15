package queue

import (
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
	Topic                string
	Brokers              []string
	Async                bool
	Username             string
	Password             string
	CA                   string
	Protocol             string
	SASLMechanism        string
	KafkaDeliveryReports bool
	KafkaProduceMaxMessages int
	KafkaQueueMaxKBytes int
	Debug				 bool
}

// Producer consumes in and produces to the topic in config
// Each message is sent to the writer via a goroutine so that the internal batch
// buffer has an opportunity to fill.
func Producer(in chan validators.ValidationMessage, config *ProducerConfig) {

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
			"queue.buffering.max.messages": config.KafkaProduceMaxMessages,
			"queue.buffering.max.kbytes": config.KafkaQueueMaxKBytes,	
		}
	} else {
		configMap = kafka.ConfigMap{
			"bootstrap.servers":   config.Brokers[0],
			"go.delivery.reports": config.KafkaDeliveryReports,
			"queue.buffering.max.messages": config.KafkaProduceMaxMessages,
			"queue.buffering.max.kbytes": config.KafkaQueueMaxKBytes,	
		}
	}

	if config.Debug {
		configMap.SetKey("debug", "protocol,broker,topic")
	}

	p, err := kafka.NewProducer(&configMap)

	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("Error creating kafka producer")
		return
	}

	defer p.Close()

	for v := range in {
		go func(v validators.ValidationMessage) {
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
			}, nil)
			messagePublishElapsed.With(prom.Labels{"topic": config.Topic}).Observe(time.Since(start).Seconds())

			e := <-p.Events()
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
