package pipeline_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "cloud.redhat.com/ingress/pipeline"
	"cloud.redhat.com/ingress/stage"
	"cloud.redhat.com/ingress/stage/local"
	"cloud.redhat.com/ingress/validators"
	"cloud.redhat.com/ingress/announcers"
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
		vch := make(chan *validators.Request)
		ach := make(chan *announcers.AvailableEvent)

		validator = &validators.Fake{
			Out: vch,
			AnnouncerChan: ach,
		}
		announcer = &announcers.Fake{}

		p = &Pipeline{
			Stager:    stager,
			Validator: validator,
			AnnouncerChan: ach,
			Announcer: announcer,
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

			vout := validator.Wait()

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

			validator.Wait()
			aout := validator.WaitForAnnounce()

			Expect(aout).To(Not(BeNil()))
			Expect(aout.URL).To(Equal(r.URL))
		})
	})
})
