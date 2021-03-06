package track

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/platform-go-middlewares/identity"

	"github.com/sirupsen/logrus"
)

// NewHandlers returns an http handler for tracking
func NewHandler(
	cfg config.IngressConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id identity.XRHID
		reqID := chi.URLParam(r, "requestID")
		verbosity := r.URL.Query().Get("verbosity")
		if verbosity == "" {
			verbosity = "0"
		}
		requestLogger := l.Log.WithFields(logrus.Fields{"source_host": cfg.Hostname, "name": "ingress"})

		logerr := func(msg string, err error) {
			requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)
		}

		if cfg.Auth == true {
			id = identity.Get(r.Context())
		}

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
		
		if id.Identity.Type != "Associate" {
			if pt.Data[0].Account != id.Identity.AccountNumber {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// Response with minimal status data by default
		latestStatus := pt.Data[len(pt.Data)-1]
		ms := MinimalStatus{
			Status: latestStatus.Status,
			Date: latestStatus.Date,
			StatusMsg: latestStatus.StatusMsg,
			Service: latestStatus.Service,
			InventoryID: latestStatus.InventoryID,
		}

		responseBody, err = json.Marshal(&ms)
		if err != nil {
			logerr("Failed to marshal JSON response", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		if verbosity >= "2" {
			w.Write(body)
		} else {
			w.Write(responseBody)
		}
	}
}