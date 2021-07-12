package track

type TrackerResponse struct {
	Data []Status `json:"data"`
	Duration interface{} `json:"duration"`
}

type Status struct {
	StatusMsg string `json:"status_msg,omitempty"`
	Date string `json:"date,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Account string `json:"account,omitempty"`
	InventoryID string `json:"inventory_id,omitempty"`
	Service string `json:"service,omitempty"`
	Status string `json:"status,omitempty"`
}

type MinimalStatus struct {
	StatusMsg string `json:"status_msg,omitempty"`
	Date string `json:"date,omitempty"`
	InventoryID string `json:"inventory_id,omitempty"`
	Service string `json:"service,omitempty"`
	Status string `json:"status,omitempty"`
}
