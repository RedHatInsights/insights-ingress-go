package kafka

import "github.com/redhatinsights/insights-ingress-go/validators"

// KafkaValidator posts requests to topics for validation
type KafkaValidator struct {
	ValidationProducerMapping map[string]chan []byte
	ValidationConsumerChannel chan []byte
	KafkaBrokers              []string
	KafkaGroupID              string
}

// ProducerConfig configures a producer
type ProducerConfig struct {
	Topic   string
	Brokers []string
}

// ConsumerConfig configures a consumer
type ConsumerConfig struct {
	Topic   string
	Brokers []string
	GroupID string
}

type KafkaConfig struct {
	Brokers         []string
	GroupID         string
	AvailableTopic  string
	ValidationTopic string
	ValidChan       chan *validators.Response
	InvalidChan     chan *validators.Response
}
