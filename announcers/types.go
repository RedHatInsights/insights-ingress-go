package announcers

import (
	"time"
)

// Announcer for messages
type Announcer interface {
	Status(e *Status)
	Stop()
}

// Status is the message sent to the payload tracker
type Status struct {
	Service     string    `json:"service"`
	Source      string    `json:"source,omitempty"`
	Account     string    `json:"account"`
	RequestID   string    `json:"request_id"`
	InventoryID string    `json:"inventory_id,omitempty"`
	SystemID    string    `json:"system_id,omitempty"`
	Status      string    `json:"status"`
	StatusMsg   string    `json:"status_msg"`
	Date        time.Time `json:"date"`
}
