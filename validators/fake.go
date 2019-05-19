package validators

import "time"

type Fake struct {
	Out chan *Request
}

func (v *Fake) Validate(in *Request) {
	v.Out <- in
}

func (v *Fake) Wait() *Request {
	select {
	case in := <-v.Out:
		return in
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}
