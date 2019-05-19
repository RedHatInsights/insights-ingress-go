package validators

// ValidationRequest is sent to the validation topic for each new payload
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
