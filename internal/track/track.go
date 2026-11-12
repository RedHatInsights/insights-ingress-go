package track

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/redhatinsights/platform-go-middlewares/v2/request_id"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	AutomatedIntegrationTestCertSubjectStage = "68bdb922-9e82-445c-a457-f83c13d23e3d"
	AutomatedIntegrationTestCertSubjectProd  = "bb2c0145-d185-4cc9-9ba1-2cdbaf90e60f"
)

type TrackerResponse struct {
	Data     []Status    `json:"data"`
	Duration interface{} `json:"duration"`
}

type Status struct {
	StatusMsg   string `json:"status_msg,omitempty"`
	Date        string `json:"date,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
	Account     string `json:"account,omitempty"`
	OrgID       string `json:"org_id,omitempty"`
	InventoryID string `json:"inventory_id,omitempty"`
	Service     string `json:"service,omitempty"`
	Status      string `json:"status,omitempty"`
}

type MinimalStatus struct {
	StatusMsg   string `json:"status_msg,omitempty"`
	Date        string `json:"date,omitempty"`
	InventoryID string `json:"inventory_id,omitempty"`
	Service     string `json:"service,omitempty"`
	Status      string `json:"status,omitempty"`
}

// NewHandlers returns an http handler for tracking
func NewHandler(
	cfg config.IngressConfig, client *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id identity.XRHID
		reqID := chi.URLParam(r, "requestID")
		requestLogger := l.Log.WithFields(logrus.Fields{"source_host": cfg.Hostname, "name": "ingress"})

		logerr := func(msg string, err error) {
			requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)
		}

		if cfg.Auth {
			id = identity.GetIdentity(r.Context())
		}

		if !isValidUUID(reqID) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("request id is not an uuid"))
			return
		}

		verbosity, _ := strconv.Atoi(r.URL.Query().Get("verbosity"))

		ptRequest, err := http.NewRequestWithContext(r.Context(), http.MethodGet, cfg.PayloadTrackerURL+reqID, nil)
		if err != nil {
			logerr("Failed to build payload-tracker request", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Forward the caller's identity so payload-tracker can authenticate
		// the request. Org-level authorization still happens below.
		if rawIdentity := forwardedIdentity(r.Context()); rawIdentity != "" {
			ptRequest.Header.Set("x-rh-identity", rawIdentity)
		}

		response, err := client.Do(ptRequest)
		if err != nil {
			logerr("Failed to get payload status", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		// Payload-tracker answers 404 for unknown request IDs; anything else
		// that is not a 200 (e.g. 401/403 from its identity check) is an
		// upstream failure, not a missing payload.
		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotFound {
			requestLogger.WithFields(logrus.Fields{"status": response.StatusCode}).Error("Unexpected status from payload-tracker")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logerr("Unable to read response body", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var pt TrackerResponse
		var responseBody []byte
		if err = json.Unmarshal(body, &pt); err != nil {
			logerr("Failed to unmarshal tracker json", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(pt.Data) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if id.Identity.Type != "Associate" && isTrustedIntegrationTestCert(id) == false {
			if !isIdAuthorized(id.Identity, pt.Data[0].OrgID) {
				incomingRequestID := request_id.GetReqID(r.Context())

				requestLogger.WithFields(logrus.Fields{"requestID": reqID,
					"incomingRequestID":         incomingRequestID,
					"pt.Data[0].Account":        pt.Data[0].Account,
					"pt.Data[0].OrgID":          pt.Data[0].OrgID,
					"id.Identity.AccountNumber": id.Identity.AccountNumber,
					"id.Identity.OrgID":         id.Identity.OrgID,
					"id.Identity.Type":          id.Identity.Type,
					"id.Identity.AuthType":      id.Identity.AuthType,
				}).Info("Returning a 403 while retrieving a payload from payload-tracker")

				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// Response with minimal status data by default
		latestStatus := pt.Data[0]
		ms := MinimalStatus{
			Status:      latestStatus.Status,
			Date:        latestStatus.Date,
			StatusMsg:   latestStatus.StatusMsg,
			Service:     latestStatus.Service,
			InventoryID: latestStatus.InventoryID,
		}

		responseBody, err = json.Marshal(&ms)
		if err != nil {
			logerr("Failed to marshal JSON response", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		if verbosity >= 2 {
			w.Write(body)
		} else {
			w.Write(responseBody)
		}
	}
}

// forwardedIdentity returns the caller's raw x-rh-identity header as stored in
// the request context by the identity middleware, re-encoding the parsed
// identity if the raw value is not available.
func forwardedIdentity(ctx context.Context) string {
	if raw := identity.GetRawIdentity(ctx); raw != "" {
		return raw
	}
	return identity.EncodeIdentity(ctx)
}

func isTrustedIntegrationTestCert(id identity.XRHID) bool {
	if id.Identity.X509 == nil {
		return false
	}

	if id.Identity.X509.SubjectDN == "" {
		return false
	}

	subjectSplit := strings.Split(id.Identity.X509.SubjectDN, "=")
	subjectDN := subjectSplit[len(subjectSplit)-1]

	return subjectDN == AutomatedIntegrationTestCertSubjectStage || subjectDN == AutomatedIntegrationTestCertSubjectProd
}

func isIdAuthorized(identity identity.Identity, orgID string) bool {
	return identity.OrgID == orgID
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
