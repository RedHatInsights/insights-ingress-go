package version

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	cfg "github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
)

func constructVersion() *IngressVersion {
	return &IngressVersion{
		Commit:  cfg.Get().OpenshiftBuildCommit,
		Version: cfg.Get().Version,
	}
}

func ExposeVersion() {
	v := constructVersion()
	versionMetric := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace:   "ingress",
		Name:        "version",
		Help:        "Version information for Ingress service.",
		ConstLabels: prometheus.Labels{"commit": v.Commit, "version": v.Version},
	})
	versionMetric.Set(1.0)
}

// GetVersion deals with calls to the version endpoint
func GetVersion(w http.ResponseWriter, r *http.Request) {
	v := constructVersion()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(v)
	if err != nil {
		l.Log.Error("Unable to get version")
		w.Write([]byte(`{"version": "unavailable"}`))
	} else {
		w.Write(jsonData)
	}
}
