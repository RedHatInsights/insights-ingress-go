package upload

// TopicDescriptor is used to select a message topic
type TopicDescriptor struct {
	Service  string `json:"service"`
	Category string `json:"category"`
}
