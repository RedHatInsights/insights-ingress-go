package validators

// ValidationRequest is sent to the validation topic for each new payload
type Request struct {
	Account     string      `json:"account"`
	B64Identity []byte      `json:"b64_identity"`
	Category    string      `json:"category"`
	Metadata    interface{} `json:"metadata"`
	PayloadID   string      `json:"payload_id"`
	Principal   string      `json:"principal"`
	Service     string      `json:"service"`
	Size        int64       `json:"size"`
	URL         string      `json:"url"`
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
