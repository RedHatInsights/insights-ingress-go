package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/sirupsen/logrus"
)

var dependencyHealth = pa.NewGaugeVec(prom.GaugeOpts{
	Name: "ingress_dependency_health",
	Help: "Health of an Ingress readiness dependency (1 is healthy, 0 is unhealthy).",
}, []string{"dependency"})

// Checker verifies that a dependency required by the upload path is usable.
type Checker interface {
	Check(context.Context) error
}

// Dependency describes one readiness dependency.
type Dependency struct {
	Name    string
	Backend string
	Checker Checker
}

type dependencyResult struct {
	Status  string `json:"status"`
	Backend string `json:"backend,omitempty"`
	Error   string `json:"error,omitempty"`
}

type response struct {
	Status       string                      `json:"status"`
	Dependencies map[string]dependencyResult `json:"dependencies"`
}

// Handler returns a dependency-aware readiness handler. Checks run in parallel
// and share one deadline so a slow dependency cannot hold a probe open forever.
func Handler(dependencies []Dependency, timeout time.Duration) http.HandlerFunc {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		type result struct {
			name   string
			status dependencyResult
		}
		results := make(chan result, len(dependencies))
		for _, dependency := range dependencies {
			dependency := dependency
			go func() {
				if dependency.Checker == nil {
					results <- result{
						name:   dependency.Name,
						status: dependencyResult{Status: "error", Backend: dependency.Backend, Error: "check is not configured"},
					}
					return
				}

				err := dependency.Checker.Check(ctx)
				if err != nil {
					l.Log.WithFields(logrus.Fields{
						"request_id": requestID(r),
						"dependency": dependency.Name,
						"error":      err,
					}).Error("Readiness dependency check failed")
					results <- result{
						name: dependency.Name,
						// The detailed error is logged above. Keep credentials and
						// endpoint details out of the probe response.
						status: dependencyResult{Status: "error", Backend: dependency.Backend, Error: "dependency check failed"},
					}
					return
				}
				dependencyHealth.WithLabelValues(dependency.Name).Set(1)
				results <- result{name: dependency.Name, status: dependencyResult{Status: "ok", Backend: dependency.Backend}}
			}()
		}

		ready := true
		dependenciesResult := make(map[string]dependencyResult, len(dependencies))
		for completed := 0; completed < len(dependencies); completed++ {
			select {
			case checkResult := <-results:
				dependenciesResult[checkResult.name] = checkResult.status
				if checkResult.status.Status != "ok" {
					ready = false
				}
			case <-ctx.Done():
				ready = false
				for _, dependency := range dependencies {
					if _, ok := dependenciesResult[dependency.Name]; !ok {
						l.Log.WithFields(logrus.Fields{
							"request_id": requestID(r),
							"dependency": dependency.Name,
							"error":      ctx.Err(),
						}).Error("Readiness dependency check timed out")
						dependenciesResult[dependency.Name] = dependencyResult{
							Status:  "error",
							Backend: dependency.Backend,
							Error:   "dependency check timed out",
						}
						dependencyHealth.WithLabelValues(dependency.Name).Set(0)
					}
				}
				completed = len(dependencies)
			}
		}

		for _, dependency := range dependencies {
			if dependenciesResult[dependency.Name].Status != "ok" {
				dependencyHealth.WithLabelValues(dependency.Name).Set(0)
			}
		}

		status := http.StatusOK
		bodyStatus := "ok"
		if !ready {
			status = http.StatusServiceUnavailable
			bodyStatus = "error"
		}
		writeJSON(w, status, response{Status: bodyStatus, Dependencies: dependenciesResult})
	}
}

func requestID(r *http.Request) string {
	return r.Header.Get("x-rh-insights-request-id")
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// CheckFunc adapts a function into a Checker.
type CheckFunc func(context.Context) error

func (f CheckFunc) Check(ctx context.Context) error {
	return f(ctx)
}
