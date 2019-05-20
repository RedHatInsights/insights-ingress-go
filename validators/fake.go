package validators

import (
	"time"
	"cloud.redhat.com/ingress/announcers"
)

type Fake struct {
	Out chan *Request
	AnnouncerChan chan *announcers.AvailableEvent
}

func (v *Fake) Validate(in *Request) {
	v.Out <- in
	v.AnnouncerChan <- &announcers.AvailableEvent{
		Account: in.Account,
		RequestID: in.RequestID,
		Principal: in.Principal,
		Service: in.Service,
		URL: in.URL,
	}
}

func (v *Fake) Wait() *Request {
	select {
	case in := <-v.Out:
		return in
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func (v *Fake) WaitForAnnounce() *announcers.AvailableEvent {
	select {
	case in := <-v.AnnouncerChan:
		return in
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}