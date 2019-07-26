package validators

import (
	"errors"
	"time"
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
	if v.DesiredResponse == "success" {
		v.Valid <- v.Out
	} else if v.DesiredResponse == "failure" {
		v.Invalid <- v.Out
	} else {
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

// WaitFor waits for a response in the channel
func (v *Fake) WaitFor(ch chan *Response) *Response {
	select {
	case o := <-ch:
		return o
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}
