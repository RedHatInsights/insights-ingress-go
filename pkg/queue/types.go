package queue

// ProducerConfig configures a producer
type ProducerConfig struct {
	Topic    string
	Brokers  []string
	Async    bool
	Username string
	Password string
	CA       string
}

// ConsumerConfig configures a consumer
type ConsumerConfig struct {
	Topic    string
	Brokers  []string
	GroupID  string
	Username string
	Password string
	CA       string
}
