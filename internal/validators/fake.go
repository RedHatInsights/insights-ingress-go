package validators

import (
	"errors"
)

// Fake allows for creation of testing objects
type Fake struct {
	In     *Request
	Called bool
	BufferFull bool
}

func (v *Fake) LoadBuffer(in *Request) error {
	v.Called = true
	v.In = in
	if v.BufferFull {
		return errors.New("buffer full")
	}
	return nil
}

// ValidateService allows for testing service validations
func (v *Fake) ValidateService(service *ServiceDescriptor) error {
	if service.Service == "failed" {
		return errors.New("failed is an invalid service")
	}
	return nil
}
