package track

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/platform-go-middlewares/identity"

	"github.com/sirupsen/logrus"
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
	cfg config.IngressConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id identity.XRHID
		reqID := chi.URLParam(r, "requestID")
		requestLogger := l.Log.WithFields(logrus.Fields{"source_host": cfg.Hostname, "name": "ingress"})

		logerr := func(msg string, err error) {
			requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)
		}

		if cfg.Auth {
			id = identity.Get(r.Context())
		}

		verbosity, _ := strconv.Atoi(r.URL.Query().Get("verbosity"))
		response, err := http.Get(cfg.PayloadTrackerURL + reqID)
		if err != nil {
			logerr("Failed to get payload status", err)
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

		var subjectDN string
		if id.Identity.X509.SubjectDN != "" {
			subjectSplit := strings.Split(id.Identity.X509.SubjectDN, "=")
			subjectDN = subjectSplit[len(subjectSplit)-1]
		}
		fmt.Print(id.Identity.Type)
		fmt.Print(subjectDN)
		if id.Identity.Type != "Associate" && subjectDN != "insightspipelineqe" {
			if !isIdAuthorized(id.Identity, pt.Data[0].Account, pt.Data[0].OrgID) {
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

func isIdAuthorized(identity identity.Identity, accountNumber string, orgID string) bool {
	return identity.AccountNumber == accountNumber || identity.OrgID == orgID
}
