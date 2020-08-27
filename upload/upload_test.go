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
	"github.com/redhatinsights/insights-ingress-go/config"
	"github.com/redhatinsights/insights-ingress-go/stage"
	. "github.com/redhatinsights/insights-ingress-go/upload"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

type FilePart struct {
	Name        string
	Content     string
	ContentType string
}

func setTime() time.Time {
	return time.Now()
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
		stager    *stage.Fake
		tracker   announcers.Announcer
		validator *validators.Fake
		handler   http.Handler
		rr        *httptest.ResponseRecorder
		timeNow   time.Time
	)

	var boiler = func(code int, parts ...*FilePart) {
		req, err := makeMultipartRequest("/api/ingress/v1/upload", parts...)
		Expect(err).To(BeNil())
		handler.ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(code))
	}

	BeforeEach(func() {

		stager = &stage.Fake{ShouldError: false}
		validator = &validators.Fake{}
		tracker = &announcers.Fake{}

		rr = httptest.NewRecorder()
		handler = NewHandler(stager, validator, tracker, *config.Get())
		timeNow = setTime()
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
				vin.Metadata.StaleTimestamp = timeNow
				Expect(vin).To(Not(BeNil()))
				Expect(vin.Metadata).To(Equal(validators.Metadata{Account: "012345", Reporter: "ingress", StaleTimestamp: timeNow}))
			})
		})

		Context("with an invalid metadata part", func() {
			It("will still return HTTP 202", func() {
				boiler(http.StatusAccepted,
					&FilePart{
						Name:        "file",
						Content:     "testing",
						ContentType: "application/vnd.redhat.unit.test",
					},
					&FilePart{
						Name:        "metadata",
						Content:     `{"account": 42}`,
						ContentType: "application/json",
					},
				)
				in := stager.Input
				Expect(in).To(Not(BeNil()))
				vin := validator.In
				Expect(vin).To(Not(BeNil()))
				Expect(vin.Metadata).To(Equal(validators.Metadata{}))
				Expect(vin.ID).To(Equal(""))
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
			It("should return HTTP 400", func() {
				boiler(http.StatusBadRequest, &FilePart{
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
				Expect(stager.StageCalled()).To(BeTrue())
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
				Expect(in).To(Not(BeNil()))
				Expect(req).To(Not(BeNil()))
				Expect(req.Service).To(Equal("unit"))
				Expect(req.Category).To(Equal("test"))
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

		Context("with alternate legacy content type", func() {
			It("should validate and be processed", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/gzip",
				})
			})
		})

		Context("with new file command legacy type", func() {
			It("should validate and be processed", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/gzip; charset=binary",
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

		Context("with content that is larger than the max allowed size", func() {
			It("should return 413", func() {
				cfg := config.Get()
				cfg.MaxSize = 1
				handler = NewHandler(stager, validator, tracker, *cfg)
				boiler(http.StatusRequestEntityTooLarge, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test",
				})
			})
		})

		Context("when the payload fails to stage", func() {
			It("should return 413", func() {
				stager = &stage.Fake{ShouldError: true}
				handler = NewHandler(stager, validator, tracker, *config.Get())
				boiler(http.StatusInternalServerError, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test",
				})
			})
		})
	})
})

var _ = Describe("NormalizeUserAgent", func() {
	Describe("when passed a support-operator agent", func() {
		It("should trim off the cluster id", func() {
			Expect(NormalizeUserAgent("support-operator/abc cluster/123")).To(Equal("support-operator/abc"))
		})
	})

	Describe("when passed a non support-operator agent", func() {
		It("should return the agent unchanged", func() {
			Expect(NormalizeUserAgent("curl/7.3.1")).To(Equal("curl/7.3.1"))
		})
	})

	Describe("when passed an insights-client agent with Core", func() {
		It("should return the client and core version should be returned", func() {
			Expect(NormalizeUserAgent("Foreman/1.22.2;redhat_access/2.2.5;insights-client/3.0.6 (Core 3.0.150; requests 2.6.0) Red Hat Enterprise Linux Server 7.7 (CPython 2.7.5; Linux 3.10.0-1062.9.1.el7.x86_64)")).To(Equal("insights-client/3.0.6 Core 3.0.150"))
		})
	})

	Describe("when passed an insights-client agent without Core", func() {
		It("should return the client version only", func() {
			Expect(NormalizeUserAgent("Satellite/6.6.2;redhat_access/2.2.8;insights-client/3.0.121")).To(Equal("insights-client/3.0.121"))
		})
	})

	Describe("when passed redhat-access-insights agents", func() {
		It("should return the client version only", func() {
			Expect(NormalizeUserAgent("Satellite/6.6.2;redhat_access/2.2.8;redhat-access-insights/1.0.13")).To(Equal("redhat-access-insights/1.0.13"))
		})
	})
})
