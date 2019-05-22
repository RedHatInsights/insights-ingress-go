package upload_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/platform-go-middlewares/identity"

	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	. "github.com/redhatinsights/insights-ingress-go/upload"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

type FakeStager struct {
	Out chan *stage.Input
}

func (s *FakeStager) Stage(in *stage.Input) (string, error) {
	s.Out <- in
	return "fake_url", nil
}

func (s *FakeStager) Reject(rawurl string) error {
	return nil
}

type FilePart struct {
	Name        string
	Content     string
	ContentType string
}

func makeMultipartRequest(uri string, parts ...*FilePart) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, filePart := range parts {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s.txt"`, filePart.Name, filePart.Name))
		h.Set("Content-Type", filePart.ContentType)

		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, strings.NewReader(filePart.Content))
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, identity.Key, identity.XRHID{
		AccountNumber: "540155",
		Internal: identity.Internal{
			OrgID: "12345",
		},
	})

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

var _ = Describe("Upload", func() {
	var (
		ch        chan *stage.Input
		reqCh     chan *validators.Request
		vCh       chan *validators.Response
		iCh       chan *validators.Response
		stager    *FakeStager
		validator *validators.Fake
		handler   http.Handler
		rr        *httptest.ResponseRecorder
		pl        *pipeline.Pipeline
	)

	var boiler = func(code int, parts ...*FilePart) {
		req, err := makeMultipartRequest("/upload", parts...)
		Expect(err).To(BeNil())
		handler.ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(code))
	}

	var waitForStager = func() *stage.Input {
		select {
		case in := <-ch:
			return in
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	var waitForValidator = func() (*validators.Request, *validators.Response) {
		var req *validators.Request
		var res *validators.Response

		fmt.Printf("About to read from reqCh: %v\n", reqCh)
		select {
		case in := <-reqCh:
			req = in
		case <-time.After(100 * time.Millisecond):
			return nil, nil
		}

		fmt.Printf("About to read from vCh: %v\n", vCh)
		select {
		case in := <-vCh:
			res = in
		case <-time.After(100 * time.Millisecond):
			return nil, nil
		}

		return req, res
	}

	BeforeEach(func() {
		ch = make(chan *stage.Input)
		reqCh = make(chan *validators.Request)
		vCh = make(chan *validators.Response)
		iCh = make(chan *validators.Response)
		stager = &FakeStager{Out: ch}
		validator = &validators.Fake{
			In:              reqCh,
			Valid:           vCh,
			Invalid:         iCh,
			DesiredResponse: "success",
		}
		pl = &pipeline.Pipeline{
			Stager:      stager,
			Validator:   validator,
			ValidChan:   vCh,
			InvalidChan: iCh,
		}
		rr = httptest.NewRecorder()
		handler = NewHandler(pl)
	})

	Describe("Posting a file to /upload", func() {
		Context("with a valid Content-Type", func() {
			It("should return HTTP 202", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test"})
			})
		})

		Context("with a metadata part", func() {
			It("should return HTTP 202", func() {
				boiler(http.StatusAccepted,
					&FilePart{
						Name:        "file",
						Content:     "testing",
						ContentType: "application/vnd.redhat.unit.test",
					},
					&FilePart{
						Name:        "metadata",
						Content:     "md",
						ContentType: "text/plain",
					},
				)
				in := waitForStager()
				Expect(in).To(Not(BeNil()))
				buf := make([]byte, 2)
				bytesRead, err := in.Metadata.Read(buf)
				Expect(err).To(BeNil())
				Expect(bytesRead).To(Equal(2))
				Expect(string(buf)).To(Equal("md"))
			})
		})

		Context("with an invalid Content-Type", func() {
			It("should return HTTP 415", func() {
				boiler(http.StatusUnsupportedMediaType, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/invalid",
				})
			})
		})

		Context("with an incorrect part name", func() {
			It("should return HTTP 415", func() {
				boiler(http.StatusUnsupportedMediaType, &FilePart{
					Name:        "invalid",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test",
				})
			})
		})

		Context("with a valid Content-Type", func() {
			It("should invoke the stager", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test"})
				Expect(waitForStager()).To(Not(BeNil()))
			})
		})

		Context("with a valid Content-Type", func() {
			It("should parse to service and category", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test"})
				in := waitForStager()
				Expect(in).To(Not(BeNil()))
				req, res := waitForValidator()
				Expect(req).To(Not(BeNil()))
				Expect(req.Service).To(Equal("unit"))
				Expect(req.Category).To(Equal("test"))
				Expect(res.Validation).To(Equal("success"))
			})
		})
	})
})
