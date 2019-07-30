package pipeline_test

import (
	"context"
	"time"

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
		r         *validators.Response
		tick      func(chan struct{})
		stopped   chan struct{}
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
		r = &validators.Response{
			Account:   "000001",
			RequestID: "testing",
		}

		p = &Pipeline{
			Stager:      stager,
			Validator:   validator,
			Announcer:   announcer,
			ValidChan:   vCh,
			InvalidChan: iCh,
			Tracker:     tracker,
		}

		stopped = make(chan struct{})

		tick = func(stopped chan struct{}) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			p.Start(ctx, stopped)
		}
	})

	Describe("A response on ValidChan", func() {
		It("should announce", func() {
			go tick(stopped)
			p.ValidChan <- r
			Expect(stager.GetURLCalled()).To(BeTrue())
			Expect(announcer.AnnounceCalled()).To(BeTrue())
		})
	})

	Describe("A failure to get a url", func() {
		It("should fail early", func() {
			failStager := &stage.Fake{ShouldError: true}
			p.Stager = failStager
			go tick(stopped)
			p.ValidChan <- r
			Expect(failStager.GetURLCalled()).To(BeTrue())
			Expect(announcer.AnnounceCalled()).To(BeFalse())
		})
	})

	Describe("A response on InvalidChan", func() {
		It("Should reject the payload", func() {
			go tick(stopped)
			p.InvalidChan <- r
			Expect(announcer.AnnounceCalled()).To(BeFalse())
			Expect(stager.RejectCalled()).To(BeTrue())
		})
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
