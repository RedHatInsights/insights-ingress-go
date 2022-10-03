package validators

import (
	"time"
)

// Request is sent to the validation topic for each new payload
type Request struct {
	Account     string    `json:"account"`
	Category    string    `json:"category"`
	ContentType string    `json:"content_type"`
	Metadata    Metadata  `json:"metadata"`
	RequestID   string    `json:"request_id"`
	Principal   string    `json:"principal"`
	OrgID       string    `json:"org_id"`
	Service     string    `json:"service"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	ID          string    `json:"id,omitempty"`
	B64Identity string    `json:"b64_identity"`
	Timestamp   time.Time `json:"timestamp"`
}

// Metadata is the expected data from a client
type Metadata struct {
	IPAddresses    []string          `json:"ip_addresses,omitempty"`
	Account        string            `json:"account,omitempty"`
	OrgID          string            `json:"org_id,omitempty"`
	InsightsID     string            `json:"insights_id,omitempty"`
	MachineID      string            `json:"machine_id,omitempty"`
	SubManID       string            `json:"subscription_manager_id,omitempty"`
	MacAddresses   []string          `json:"mac_addresses,omitempty"`
	FQDN           string            `json:"fqdn,omitempty"`
	BiosUUID       string            `json:"bios_uuid,omitempty"`
	DisplayName    string            `json:"display_name,omitempty"`
	AnsibleHost    string            `json:"ansible_host,omitempty"`
	CustomMetadata map[string]string `json:"custom_metadata,omitempty"`
	Reporter       string            `json:"reporter"`
	StaleTimestamp time.Time         `json:"stale_timestamp"`
	QueueKey       string            `json:"queue_key,omitempty"`
}

type ValidationMessage struct {
	Message []byte
	Headers map[string]string
	Key     []byte
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
