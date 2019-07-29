package validators

import (
	"errors"
)

// Fake allows for creation of testing objects
type Fake struct {
	In              *Request
	Out             *Response
	Valid           chan *Response
	Invalid         chan *Response
	Called          bool
	DesiredResponse string
}

// Validate creates a fake validation response
func (v *Fake) Validate(in *Request) {
	v.Called = true
	v.In = in
	v.Out = &Response{
		RequestID:  in.RequestID,
		Validation: v.DesiredResponse,
		URL:        in.URL,
		Account:    in.Account,
		Principal:  in.Principal,
		Service:    in.Service,
	}
	switch v.DesiredResponse {
	case "success":
		v.Valid <- v.Out
	case "failure":
		v.Invalid <- v.Out
	default:
		return
	}
}

// ValidateService allows for testing service validations
func (v *Fake) ValidateService(service *ServiceDescriptor) error {
	if service.Service == "failed" {
		return errors.New("failed is an invalid service")
	}
	return nil
}
