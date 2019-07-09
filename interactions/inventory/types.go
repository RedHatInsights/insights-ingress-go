package inventory

import "github.com/redhatinsights/insights-ingress-go/validators"

// Response is the JSON stucture of the inventory response
type Response struct {
	Data []struct {
		Status int `json:"status"`
		Host   struct {
			ID string `json:"id"`
		} `json:"host"`
	} `json:"data"`
}

// Inventory can return an inventory ID
type Inventory interface {
	GetID(vr *validators.Request) (string, error)
}
