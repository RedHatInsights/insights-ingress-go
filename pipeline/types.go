package pipeline

import (
	"cloud.redhat.com/ingress/announcers"
	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/validators"
)

// Pipeline defines the descrete processing steps for ingress
type Pipeline struct {
	Stager      stage.Stager
	Validator   validators.Validator
	Announcer   announcers.Announcer
	ValidChan   chan *validators.Response
	InvalidChan chan *validators.Response
}
