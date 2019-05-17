package upload_test

import (
	"bytes"
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

	. "cloud.redhat.com/ingress/upload"
)

type FakeStager struct {
	Out chan *StageInput
}

func (s *FakeStager) Stage(in *StageInput) (string, error) {
	s.Out <- in
	return "", nil
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

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

var _ = Describe("Upload", func() {
	var (
		ch      chan *StageInput
		stager  *FakeStager
		handler http.Handler
		rr      *httptest.ResponseRecorder
	)

	var boiler = func(code int, parts ...*FilePart) {
		req, err := makeMultipartRequest("/upload", parts...)
		Expect(err).To(BeNil())
		handler.ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(code))
	}

	var waitForStager = func() *StageInput {
		select {
		case in := <-ch:
			return in
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	BeforeEach(func() {
		ch = make(chan *StageInput)
		stager = &FakeStager{Out: ch}
		rr = httptest.NewRecorder()
		handler = NewHandler(stager)
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

		Context("with a valid Content-Type", func() {
			It("should invoke the stager", func() {
				boiler(http.StatusAccepted, &FilePart{
					Name:        "file",
					Content:     "testing",
					ContentType: "application/vnd.redhat.unit.test"})
				Expect(waitForStager()).To(Not(BeNil()))
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
	})
})
