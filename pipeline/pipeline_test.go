package pipeline_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	)

	BeforeEach(func() {
		stager = local.New("/tmp")
		vch := make(chan *validators.Request)
		validator = &validators.Fake{
			Out: vch,
		}
		p = &Pipeline{
			Stager:    stager,
			Validator: validator,
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

			out := validator.Wait()

			Expect(r.URL).To(Not(BeNil()))
			Expect(out).To(Not(BeNil()))
			Expect(out.URL).To(Equal(r.URL))
		})
	})
})
