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
	Out chan int
}

func (s *FakeStager) Stage(file io.Reader, key string) (string, error) {
	s.Out <- 1
	return "", nil
}

func makeMultipartRequest(name string, content string, contentType string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s.txt"`, name, name))
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

var _ = Describe("Upload", func() {
	var (
		ch           chan int
		stager       *FakeStager
		handler      http.Handler
		rr           *httptest.ResponseRecorder
		stagerCalled bool
	)

	BeforeEach(func() {
		ch = make(chan int)
		stager = &FakeStager{Out: ch}
		rr = httptest.NewRecorder()
		handler = NewHandler(stager)
	})

	Describe("Posting a file to /upload", func() {
		Context("with a valid Content-Type", func() {
			It("should return HTTP 202", func() {
				req, err := makeMultipartRequest("file", "testing", "application/vnd.redhat.unit.test")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(http.StatusAccepted))
			})
		})

		Context("with a valid Content-Type", func() {
			It("should invoke the stager", func() {
				req, err := makeMultipartRequest("file", "testing", "application/vnd.redhat.unit.test")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				select {
				case <-ch:
					stagerCalled = true
				case <-time.After(100 * time.Millisecond):
					stagerCalled = false
				}
				Expect(stagerCalled).To(BeTrue())
			})
		})

		Context("with an invalid Content-Type", func() {
			It("should return HTTP 415", func() {
				req, err := makeMultipartRequest("file", "testing", "application/invalid")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(http.StatusUnsupportedMediaType))
			})
		})
	})
})
