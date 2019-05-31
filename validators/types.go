package validators

import (
	"io"
	"time"
)

// Request is sent to the validation topic for each new payload
type Request struct {
	Account     string    `json:"account"`
	Category    string    `json:"category"`
	Metadata    io.Reader `json:"metadata"`
	RequestID   string    `json:"request_id"`
	Principal   string    `json:"principal"`
	Service     string    `json:"service"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	B64Identity string    `json:"b64_identity"`
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
}

// Response is returned by validators and sent via the announcement
type Response struct {
	Account     string            `json:"account"`
	Validation  string            `json:"validation"`
	RequestID   string            `json:"request_id"`
	Principal   string            `json:"principal"`
	Service     string            `json:"service"`
	URL         string            `json:"url"`
	B64Identity string            `json:"b64_identity"`
	Extras      map[string]string `json:"extras"`
	Timestamp   time.Time         `json:"timestamp"`
}

// ServiceDescriptor is used to select a message topic
type ServiceDescriptor struct {
	Service  string
	Category string
}

// Validator validates requests
type Validator interface {
	Validate(req *Request)
	ValidateService(service *ServiceDescriptor) error
}
