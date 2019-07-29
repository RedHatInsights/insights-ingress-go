package pipeline_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhatinsights/insights-ingress-go/announcers"
	. "github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

var _ = Describe("Pipeline", func() {

	var (
		p         *Pipeline
		validator *validators.Fake
		stager    *stage.Fake
		announcer *announcers.Fake
		tracker   *announcers.Fake
	)

	BeforeEach(func() {
		stager = &stage.Fake{}
		vCh := make(chan *validators.Response)
		iCh := make(chan *validators.Response)

		validator = &validators.Fake{
			Valid:           vCh,
			Invalid:         iCh,
			DesiredResponse: "success",
		}
		announcer = &announcers.Fake{}
		tracker = &announcers.Fake{}

		p = &Pipeline{
			Stager:      stager,
			Validator:   validator,
			Announcer:   announcer,
			ValidChan:   vCh,
			InvalidChan: iCh,
			Tracker:     tracker,
		}
	})

	Describe("Canceling the context", func() {
		It("Should stop the Start() loop", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			stopped := make(chan struct{})
			p.Start(ctx, stopped)
			_, ok := <-stopped
			Expect(ok).To(BeFalse())
		})
	})

	Describe("Closing the valid channel", func() {
		It("should stop the Start() loop", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stopped := make(chan struct{})
			close(p.ValidChan)
			p.Start(ctx, stopped)
			_, ok := <-stopped
			Expect(ok).To(BeFalse())
		})
	})

	Describe("Closing the invalid channel", func() {
		It("should stop the Start() loop", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stopped := make(chan struct{})
			close(p.InvalidChan)
			p.Start(ctx, stopped)
			_, ok := <-stopped
			Expect(ok).To(BeFalse())
		})
	})
})
