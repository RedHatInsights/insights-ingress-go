package validators

import "cloud.redhat.com/ingress/announcers"

// Request is sent to the validation topic for each new payload
type Request struct {
	Account     string      `json:"account"`
	B64Identity []byte      `json:"b64_identity"`
	Category    string      `json:"category"`
	Metadata    interface{} `json:"metadata"`
	RequestID   string      `json:"request_id"`
	Principal   string      `json:"principal"`
	Service     string      `json:"service"`
	Size        int64       `json:"size"`
	URL         string      `json:"url"`
}

type Response struct {
	RequestID  string `json:"request_id"`
	Validation string `json:"validation"`
	URL        string `json:"url"`
	Account    string `json:"account"`
	Principal  string `json:"principal"`
	Service    string `json:"service"`
}

// Validator validates requests
type Validator interface {
	Validate(req *Request)
}

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
	Brokers []string
	GroupID string
	AvailableTopic string
	ValidationTopic string
	AnnouncerChan chan *announcers.AvailableEvent
}