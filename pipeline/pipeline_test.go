package pipeline_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"cloud.redhat.com/ingress/announcers"
	. "cloud.redhat.com/ingress/pipeline"
	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/stage/local"
	"cloud.redhat.com/ingress/validators"
)

var _ = Describe("Pipeline", func() {

	var (
		p         *Pipeline
		validator *validators.Fake
		stager    *local.LocalStager
		announcer *announcers.Fake
	)

	BeforeEach(func() {
		stager = local.New("/tmp")
		reqCh := make(chan *validators.Request)
		vCh := make(chan *validators.Response)
		iCh := make(chan *validators.Response)

		validator = &validators.Fake{
			In:              reqCh,
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

	AfterEach(func() {
		stager.CleanUp()
	})

	Describe("Submitting a valid stage.Input", func() {
		It("should return a URL", func() {
			stageIn := &stage.Input{
				Reader: strings.NewReader("test"),
			}
			r := &validators.Request{
				Account:   "123",
				RequestID: "foo",
			}
			go p.Submit(stageIn, r)

			vout := validator.WaitForIn()

			Expect(r.URL).To(Not(BeNil()))
			Expect(vout).To(Not(BeNil()))
			Expect(vout.URL).To(Equal(r.URL))
		})
	})

	Describe("Submitting a valid stage.Input", func() {
		It("should validate", func() {
			stageIn := &stage.Input{
				Reader: strings.NewReader("test"),
			}
			r := &validators.Request{
				Account:   "123",
				RequestID: "foo",
			}
			go p.Submit(stageIn, r)

			validator.WaitForIn()
			aout := validator.WaitFor(validator.Valid)

			Expect(aout).To(Not(BeNil()))
			Expect(aout.URL).To(Equal(r.URL))
		})
	})

	Describe("Submitting a payload that fails to validate", func() {
		It("should call stager.Reject", func() {
			stageIn := &stage.Input{
				Reader: strings.NewReader("invalid"),
			}
			r := &validators.Request{
				Account:   "123",
				RequestID: "invalid",
			}
			validator.DesiredResponse = "failure"
			go p.Submit(stageIn, r)

			_ = validator.WaitForIn()

			aout := validator.WaitFor(validator.Invalid)
			Expect(aout).To(Not(BeNil()))
		})
	})
})
