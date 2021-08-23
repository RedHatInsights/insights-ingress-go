package kafka

// Validator posts requests to topics for validation
type Validator struct {
	ValidationProducerMapping map[string]chan []byte
	KafkaBrokers              []string
	KafkaGroupID              string
	Username                  string
	Password                  string
	CA                        string
	SASLMechanism             string
	Protocol                  string
}

// Config configures a new Kafka Validator
type Config struct {
	Brokers         []string
	GroupID         string
	ValidationTopic string
	Username        string
	Password        string
	CA              string
	Protocol        string
	SASLMechanism   string
}
