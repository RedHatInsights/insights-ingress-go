package pipeline

import (
	"log"

	"cloud.redhat.com/ingress/stage"
)

// Validate validates a ValidationRequest
func (kv *KafkaValidator) Validate(vr *ValidationRequest) {
}

// Submit accepts a stage request and a validation request
func (p *Pipeline) Submit(in *stage.Input, vr *ValidationRequest) {
	url, err := p.Stager.Stage(in)
	if err != nil {
		log.Printf("Error staging %v: %v", in, err)
		return
	}
	vr.URL = url
	p.Validator.Validate(vr)
}

func (p *Pipeline) run() {
	for {

	}
}
