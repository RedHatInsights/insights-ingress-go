package queue

// ProducerConfig configures a producer
type ProducerConfig struct {
	Topic   string
	Brokers []string
	Async   bool
}

// ConsumerConfig configures a consumer
type ConsumerConfig struct {
	Topic   string
	Brokers []string
	GroupID string
}
