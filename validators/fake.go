package validators

import (
	"time"
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

func (v *Fake) WaitFor(ch chan *Response) *Response {
	select {
	case o := <-ch:
		return o
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}
