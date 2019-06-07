package pipeline_test

import (
	"context"
	"io/ioutil"
	"strings"

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

		p = &Pipeline{
			Stager:      stager,
			Validator:   validator,
			Announcer:   announcer,
			ValidChan:   vCh,
			InvalidChan: iCh,
		}
	})

	Describe("Submitting a valid stage.Input", func() {
		It("should return a URL", func() {
			stageIn := &stage.Input{
				Payload: ioutil.NopCloser(strings.NewReader("test")),
			}
			r := &validators.Request{
				Account:   "123",
				RequestID: "foo",
			}
			go p.Submit(stageIn, r)

			vout := validator.WaitFor(p.ValidChan)

			Expect(r.URL).To(Not(BeNil()))
			Expect(vout).To(Not(BeNil()))
			Expect(vout.URL).To(Equal(r.URL))
		})
	})

	Describe("Submitting a valid stage.Input", func() {
		It("should validate", func() {
			stageIn := &stage.Input{
				Payload: ioutil.NopCloser(strings.NewReader("test")),
			}
			r := &validators.Request{
				Account:   "123",
				RequestID: "foo",
			}
			go p.Submit(stageIn, r)

			p.Tick(context.Background())

			aout := announcer.Event

			Expect(aout).To(Not(BeNil()))
			Expect(aout.URL).To(Equal(r.URL))
		})
	})

	Describe("Submitting a payload that fails to validate", func() {
		It("should call stager.Reject", func() {
			stageIn := &stage.Input{
				Payload: ioutil.NopCloser(strings.NewReader("invalid")),
			}
			r := &validators.Request{
				Account:   "123",
				RequestID: "invalid",
			}
			validator.DesiredResponse = "failure"
			go p.Submit(stageIn, r)

			p.Tick(context.Background())

			Expect(stager.RejectCalled).To(BeTrue())
		})
	})

	Describe("An error during stage", func() {
		It("should not return a URL", func() {
			stager.ShouldError = true
			p.Stager = stager
			stageIn := &stage.Input{}
			r := &validators.Request{
				Account:   "123",
				RequestID: "test",
			}
			p.Submit(stageIn, r)
			Expect(validator.Called).To(BeFalse())
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
