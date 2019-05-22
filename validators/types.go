package validators

// Request is sent to the validation topic for each new payload
type Request struct {
	Account   string      `json:"account"`
	Category  string      `json:"category"`
	Metadata  interface{} `json:"metadata"`
	RequestID string      `json:"request_id"`
	Principal string      `json:"principal"`
	Service   string      `json:"service"`
	Size      int64       `json:"size"`
	URL       string      `json:"url"`
}

// Response is returned by validators and sent via the announcement
type Response struct {
	Account    string            `json:"account"`
	Validation string            `json:"validation"`
	RequestID  string            `json:"request_id"`
	Principal  string            `json:"principal"`
	Service    string            `json:"service"`
	URL        string            `json:"url"`
	Extras     map[string]string `json:"extras"`
}

// Validator validates requests
type Validator interface {
	Validate(req *Request)
}
