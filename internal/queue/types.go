package queue

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
}
