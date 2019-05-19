package pipeline

import (
	"log"

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

// Start enables the consumer loop and watches for validation responses
func (p *Pipeline) Start() {
	for {

	}
}
