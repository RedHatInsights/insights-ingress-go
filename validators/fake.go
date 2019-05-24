package validators

import (
	"time"
	"errors"
)

type Fake struct {
	In              *Request
	Out             *Response
	Valid           chan *Response
	Invalid         chan *Response
	Called          bool
	DesiredResponse string
}

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

func (v *Fake) ValidateService(service *ServiceDescriptor) error {
	if service.Service == "failed" {
		return errors.New("failed is an invalid service")
	}
	return nil
}

func (v *Fake) WaitFor(ch chan *Response) *Response {
	select {
	case o := <-ch:
		return o
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

type Simulation struct {
	CallDelay   time.Duration
	Delay       time.Duration
	ValidChan   chan *Response
	InvalidChan chan *Response
}

func (s *Simulation) Validate(request *Request) {
	go func() {
		time.Sleep(s.Delay)
		s.ValidChan <- &Response{
			Account:     request.Account,
			Validation:  "success",
			RequestID:   request.RequestID,
			Principal:   request.Principal,
			Service:     request.Service,
			URL:         request.URL,
			B64Identity: request.B64Identity,
		}
	}()
	time.Sleep(s.CallDelay)
}

func (s *Simulation) ValidateService(service *ServiceDescriptor) error {
	return nil
}