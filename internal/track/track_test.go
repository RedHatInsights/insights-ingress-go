package track_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"

	. "github.com/redhatinsights/insights-ingress-go/internal/track"
)

func makeTestRequest(uri string, request_id string, account string, orgID string, account_type string) (*http.Request, error) {

	var req *http.Request
	var err error

	req, err = http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("requestID", request_id)

	ctx := context.Background()

	switch account_type {
	case "associate":
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				AccountNumber: account,
				Internal: identity.Internal{
					OrgID: orgID,
				},
				Type: "Associate",
			},
		})
	case "x509":
		x509 := identity.X509{
			SubjectDN: "/DC=com/DC=redhat/CN=" + AutomatedIntegrationTestCertSubjectStage,
			IssuerDN:  "CN=redhat",
		}
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				X509: &x509,
			},
		})
	case "untrusted_x509":
		x509 := identity.X509{
			SubjectDN: "/DC=com/DC=redhat/CN=im_not_to_be_trusted",
			IssuerDN:  "CN=redhat",
		}
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				X509: &x509,
			},
		})
	case "serviceaccount":
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				OrgID: orgID,
				Internal: identity.Internal{
					OrgID: orgID,
				},
				Type: "ServiceAccount",
			},
		})
	default:
		ctx = identity.WithIdentity(ctx, identity.XRHID{
			Identity: identity.Identity{
				AccountNumber: account,
				OrgID:         orgID,
				Internal: identity.Internal{
					OrgID: orgID,
				},
			},
		})

	}

	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	return req, nil

}

var _ = Describe("Track", func() {
	var (
		handler         http.Handler
		rr              *httptest.ResponseRecorder
		goodJsonBody    = `{"data":[{"id":7738152,"status_msg":"message received","date":"2021-06-11 20:12:40.228334+00:00","created_at":"2021-06-11 20:12:40.375706+00:00","request_id":"747c1300667b441e8e8f448337588ec0","account":"6089719","org_id":"12345","inventory_id":"766637dd-653f-4bf0-99a1-a117b455cd96","system_id":"b1596dc8-fb46-4e16-8790-d11ea7dfa16a","service":"inventory-mq-service","status":"received"},{"id":7738150,"status_msg":"Message sent to inventory","date":"2021-06-11 20:12:40.212302+00:00","created_at":"2021-06-11 20:12:40.269223+00:00","request_id":"747c1300667b441e8e8f448337588ec0","account":"6089719","org_id":"12345","inventory_id":"766637dd-653f-4bf0-99a1-a117b455cd96","system_id":"b1596dc8-fb46-4e16-8790-d11ea7dfa16a","service":"puptoo","status":"success"},{"id":7738149,"status_msg":"Received message","date":"2021-06-11 20:12:39.219543+00:00","created_at":"2021-06-11 20:12:39.303714+00:00","request_id":"747c1300667b441e8e8f448337588ec0","account":"6089719","org_id":"12345","inventory_id":"766637dd-653f-4bf0-99a1-a117b455cd96","system_id":"b1596dc8-fb46-4e16-8790-d11ea7dfa16a","service":"puptoo","status":"received"}],"duration":{"hsp-archiver:undefined":"0:00:00.067749","storage-broker:undefined":"0:00:00","puptoo:undefined":"0:00:00.992759","inventory-mq-service:undefined":"0:00:00.102472","vulnerability-rules:undefined":"0:00:01.155458","insights-engine:undefined":"0:00:01.691324","vulnerability-vmaas:undefined":"0:00:02.812779","insights-advisor-service:insights-client":"0:00:00.066972","total_time_in_services":"0:00:06.889513","total_time":"0:00:04.038974"}}`
		minimalJsonBody = `{"status_msg":"message received","date":"2021-06-11 20:12:40.228334+00:00","inventory_id":"766637dd-653f-4bf0-99a1-a117b455cd96","service":"inventory-mq-service","status":"received"}`
		badJsonID       = `{"data": [], "duration": {}}`
	)

	BeforeEach(func() {

		rr = httptest.NewRecorder()
		httpClient := &http.Client{}
		handler = NewHandler(*config.Get(), httpClient)
		httpmock.Activate()
	})

	Describe("Get request id", func() {
		Context("with a valid request-id", func() {
			It("should return HTTP 200", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/3e3f56e642a248008811cce123b2c0f2", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089719", "12345", "customer")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body).ToNot(BeNil())
				Expect(rr.Body.String()).To(Equal(minimalJsonBody))
			})
		})

		Context("with a valid request-id and higher verbosity", func() {
			It("should return HTTP 200", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/3e3f56e642a248008811cce123b2c0f2", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2?verbosity=2", "3e3f56e642a248008811cce123b2c0f2", "6089719", "12345", "customer")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body).ToNot(BeNil())
				Expect(rr.Body.String()).To(Equal(goodJsonBody))
			})
		})

		Context("with an incorrect account but valid orgID", func() {
			It("should return an HTTP 200", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "", "12345", "customer")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
			})
		})

		Context("with an invalid request-id", func() {
			It("should return HTTP 404", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/3e3f56e642a248008811cce123b2c0f2", httpmock.NewStringResponder(200, badJsonID))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089719", "12345", "customer")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(404))
			})
		})

		Context("with an incorrect account and orgID", func() {
			It("should return an HTTP 403", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/3e3f56e642a248008811cce123b2c0f2", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089710", "12346", "customer")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(403))
			})
		})

		Context("with an associate login", func() {
			It("should return HTTP 200", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089710", "12346", "associate")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body).ToNot(BeNil())
				Expect(rr.Body.String()).To(Equal(minimalJsonBody))
			})
		})

		Context("with a qe test account", func() {
			It("should return HTTP 200", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089710", "12346", "x509")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body).ToNot(BeNil())
				Expect(rr.Body.String()).To(Equal(minimalJsonBody))
			})
		})

		Context("with an service-account", func() {
			It("should return HTTP 200", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089710", "12345", "serviceaccount")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(200))
				Expect(rr.Body).ToNot(BeNil())
				Expect(rr.Body.String()).To(Equal(minimalJsonBody))
			})
		})

		Context("with an service-account with an incorrect orgID", func() {
			It("should return HTTP 403", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089710", "12346", "serviceaccount")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(403))
			})
		})

		Context("with an untrusted cert", func() {
			It("should return HTTP 403", func() {
				httpmock.RegisterResponder("GET", "http://payload-tracker/v1/payloads/", httpmock.NewStringResponder(200, goodJsonBody))
				req, err := makeTestRequest("/api/ingress/v1/track/3e3f56e642a248008811cce123b2c0f2", "3e3f56e642a248008811cce123b2c0f2", "6089710", "12346", "untrusted_x509")
				Expect(err).To(BeNil())
				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(403))
				Expect(rr.Body).ToNot(BeNil())
				Expect(rr.Body.String()).To(Equal(""))
			})
		})

	})
})
