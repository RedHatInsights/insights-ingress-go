package kafka

import "github.com/redhatinsights/insights-ingress-go/validators"

// Validator posts requests to topics for validation
type Validator struct {
	ValidationProducerMapping map[string]chan []byte
	ValidationConsumerChannel chan []byte
	KafkaBrokers              []string
	KafkaGroupID              string
}

// Config configures a new Kafka Validator
type Config struct {
	Brokers         []string
	GroupID         string
	ValidationTopic string
	ValidChan       chan *validators.Response
	InvalidChan     chan *validators.Response
}
