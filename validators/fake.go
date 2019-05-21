package validators

import (
	"fmt"
	"time"
)

type Fake struct {
	In              chan *Request
	Valid           chan *Response
	Invalid         chan *Response
	DesiredResponse string
}

func (v *Fake) Validate(in *Request) {
	fmt.Println("About to push to v.In")
	v.In <- in
	r := &Response{
		RequestID:  in.RequestID,
		Validation: v.DesiredResponse,
		URL:        in.URL,
		Account:    in.Account,
		Principal:  in.Principal,
		Service:    in.Service,
	}
	if v.DesiredResponse == "success" {
		fmt.Println("About to push to v.Valid")
		v.Valid <- r
	} else if v.DesiredResponse == "failure" {
		fmt.Println("About to push to v.Invalid")
		v.Invalid <- r
	} else {
		return
	}
}

func (v *Fake) WaitForIn() *Request {
	select {
	case o := <-v.In:
		return o
	case <-time.After(100 * time.Millisecond):
		return nil
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
