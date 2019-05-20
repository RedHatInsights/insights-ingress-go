package pipeline

import (
	"log"
	"context"

	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/validators"
)

// Submit accepts a stage request and a validation request
func (p *Pipeline) Submit(in *stage.Input, vr *validators.Request) {
	url, err := p.Stager.Stage(in)
	if err != nil {
		log.Printf("Error staging %v: %v", in, err)
		return
	}
	vr.URL = url
	p.Validator.Validate(vr)
}

// Start watches the announcer channel for new events and calls announce
func (p *Pipeline) Start(ctx context.Context) {
	for {
		select {
		case ev := <- p.AnnouncerChan:
			p.Announcer.Announce(ev)
		case <-ctx.Done():
			return
		}
	}
}