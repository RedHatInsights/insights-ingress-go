package pipeline

import (
	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/validators"
	"cloud.redhat.com/ingress/announcers"
)

// Pipeline defines the descrete processing steps for ingress
type Pipeline struct {
	Stager    stage.Stager
	Validator validators.Validator
	Announcer announcers.Announcer
	AnnouncerChan chan *announcers.AvailableEvent
}
