package kafka

// Validator posts requests to topics for validation
type Validator struct {
	ValidationProducerMapping map[string]chan []byte
	KafkaBrokers              []string
	KafkaGroupID              string
}

// Config configures a new Kafka Validator
type Config struct {
	Brokers         []string
	GroupID         string
	ValidationTopic string
}
