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
	ID          string    `json:"id,omitempty"`
	B64Identity string    `json:"b64_identity"`
	Timestamp   time.Time `json:"timestamp"`
}

// Response is returned by validators and sent via the announcement
type Response struct {
	Account          string                 `json:"account"`
	Validation       string                 `json:"validation"`
	RequestID        string                 `json:"request_id"`
	Principal        string                 `json:"principal"`
	Service          string                 `json:"service"`
	URL              string                 `json:"url"`
	B64Identity      string                 `json:"b64_identity"`
	ID               string                 `json:"id,omitempty"`
	SatelliteManaged *bool                  `json:"satellite_managed,omitemtpy"`
	Extras           map[string]interface{} `json:"extras"`
	Timestamp        time.Time              `json:"timestamp"`
}

// Status is the message sent to the payload tracker
type Status struct {
	Service     string `json:"service"`
	Source      string `json:"source,omitempty"`
	Account     string `json:"account"`
	RequestID   string `json:"request_id"`
	InventoryID string `json:"inventory_id"`
	SystemID    string `json:"system_id"`
	Status      string `json:"status"`
	StatusMsg   string `json:"status_msg"`
	Date        string `json:"date"`
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
