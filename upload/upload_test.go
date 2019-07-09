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

	"github.com/redhatinsights/insights-ingress-go/announcers"
	"github.com/redhatinsights/insights-ingress-go/interactions/inventory"
	"github.com/redhatinsights/insights-ingress-go/pipeline"
	"github.com/redhatinsights/insights-ingress-go/stage"
	. "github.com/redhatinsights/insights-ingress-go/upload"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

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
		Identity: identity.Identity{
			AccountNumber: "540155",
			Internal: identity.Internal{
				OrgID: "12345",
			},
		},
	})

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

var _ = Describe("Upload", func() {
	var (
		vCh       chan *validators.Response
		iCh       chan *validators.Response
		stager    *stage.Fake
		validator *validators.Fake
		handler   http.Handler
		rr        *httptest.ResponseRecorder
		pl        *pipeline.Pipeline
	)

	var boiler = func(code int, parts ...*FilePart) {
		req, err := makeMultipartRequest("/api/ingress/v1/upload", parts...)
		Expect(err).To(BeNil())
		handler.ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(code))
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		pl.Tick(ctx)
	}

	BeforeEach(func() {
		vCh = make(chan *validators.Response)
		iCh = make(chan *validators.Response)
		stager = &stage.Fake{ShouldError: false}
		validator = &validators.Fake{
			Valid:           vCh,
			Invalid:         iCh,
			DesiredResponse: "success",
		}
		pl = &pipeline.Pipeline{
			Stager:      stager,
			Validator:   validator,
			Announcer:   &announcers.Fake{},
			ValidChan:   vCh,
			InvalidChan: iCh,
			Inventory:   &inventory.Fake{},
			Tracker:     &announcers.Fake{},
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
						Content:     `{"account": "012345"}`,
						ContentType: "text/plain",
					},
				)
				in := stager.Input
				Expect(in).To(Not(BeNil()))
				vin := validator.In
				Expect(vin).To(Not(BeNil()))
				Expect(vin.Metadata).To(Equal(validators.Metadata{Account: "012345"}))
				Expect(vin.ID).To(Equal("1234-abcd-5678-efgh"))
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

		Context("with a valid file part", func() {
			It("should return a 202", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "upload",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test",
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
				Expect(stager.StageCalled).To(BeTrue())
			})
		})

		Context("with a valid Content-Type", func() {
			It("should parse to service and category", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test"})
				in := stager.Input
				req := validator.In
				res := validator.Out
				Expect(in).To(Not(BeNil()))
				Expect(req).To(Not(BeNil()))
				Expect(req.Service).To(Equal("unit"))
				Expect(req.Category).To(Equal("test"))
				Expect(res.Validation).To(Equal("success"))
			})
		})

		Context("with legacy content type", func() {
			It("should validate and be processed", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/x-gzip; charset=binary",
				})
			})
		})

		Context("with invalid service name", func() {
			It("should return 415", func() {
				boiler(http.StatusUnsupportedMediaType, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.failed.test"})
			})
		})
	})
})
