package filebased_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"

	"github.com/redhatinsights/insights-ingress-go/internal/announcers"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/stage/filebased"
	. "github.com/redhatinsights/insights-ingress-go/internal/upload"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
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
	requestId := "e6b06142958942139a5e1e2f513c448b"
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
	ctx = identity.WithIdentity(ctx, identity.XRHID{
		Identity: identity.Identity{
			AccountNumber: "540155",
			OrgID:         "12345",
			Internal: identity.Internal{
				OrgID: "12345",
			},
		},
	})

	req.Header.Add("x-rh-insights-request-id", requestId)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func makeTestRequest(uri string, testType string, tenant string, body string) (*http.Request, error) {

	var req *http.Request
	var err error

	requestId := "e6b06142958942139a5e1e2f513c448b"

	if testType == "new" {
		formData := url.Values{"test": {"test"}}
		req, err = http.NewRequest("POST", uri, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	}

	if testType == "legacy" {
		req, err = http.NewRequest("POST", uri, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	}

	ctx := context.Background()
	if tenant == "anemic" {
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				OrgID: "12345",
				Internal: identity.Internal{
					OrgID: "12345",
				},
			},
		})
	} else {
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				AccountNumber: "540155",
				OrgID:         "12345",
				Internal: identity.Internal{
					OrgID: "12345",
				},
			},
		})
	}

	req.Header.Add("x-rh-insights-request-id", requestId)
	req = req.WithContext(ctx)
	return req, nil

}

var _ = Describe("Upload", func() {
	var (
		configuration *config.IngressConfig
		stager        *filebased.FileBasedStager
		tracker       announcers.Announcer
		validator     *validators.Fake
		handler       http.Handler
		rr            *httptest.ResponseRecorder

		goodJsonBody       = `{"request_id":"e6b06142958942139a5e1e2f513c448b","upload":{"account_number":"540155","org_id":"12345"}}`
		goodAnemicJsonBody = `{"request_id":"e6b06142958942139a5e1e2f513c448b","upload":{"org_id":"12345"}}`
	)

	BeforeEach(func() {

		configuration = config.Get()
		stager = &filebased.FileBasedStager{StagePath: configuration.StorageConfig.StorageFileSystemPath, BaseURL: configuration.ServiceBaseURL}
		validator = &validators.Fake{}
		tracker = &announcers.Fake{}

		rr = httptest.NewRecorder()
		handler = NewHandler(stager, validator, tracker, *configuration)
		reqConfiguredHandlerFunc := request_id.ConfiguredRequestID("x-rh-insights-request-id")
		handler = reqConfiguredHandlerFunc(handler)
	})

	Describe("Post to endpoint", func() {
		Context("with test data for test-connection", func() {
			It("should return HTTP 200", func() {
				req, err := makeTestRequest("/api/ingress/v1/upload", "new", "", "")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body.String()).To(Equal(goodJsonBody))
			})
		})

		Context("with missing account number in identity header", func() {
			It("should return HTTP 200", func() {
				req, err := makeTestRequest("/api/ingress/v1/upload", "new", "anemic", "")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body.String()).To(Equal(goodAnemicJsonBody))
			})
		})
	})
})
