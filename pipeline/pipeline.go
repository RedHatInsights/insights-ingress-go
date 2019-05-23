package pipeline

import (
	"context"
	"log"

	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
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

// Tick handles one loop for handling post-validation activities
func (p *Pipeline) Tick(ctx context.Context) bool {
	select {
	case ev := <-p.ValidChan:
		p.Announcer.Announce(ev)
	case iev := <-p.InvalidChan:
		p.Stager.Reject(iev.URL)
	case <-ctx.Done():
		return false
	}
	return true
}

// Start watches the announcer channel for new events and calls announce
func (p *Pipeline) Start(ctx context.Context) {
	keepGoing := true
	for keepGoing {
		keepGoing = p.Tick(ctx)
	}
}
