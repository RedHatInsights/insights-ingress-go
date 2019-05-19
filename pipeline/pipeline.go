package pipeline

import (
	"log"

	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/validators"
)

// NewKafkaValidator constructs and initializes a new Kafka Validator

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

func (p *Pipeline) run() {
	for {

	}
}
