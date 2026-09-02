package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/sirupsen/logrus"
)

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

type checkResult struct {
	name   string
	status dependencyResult
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

		results := make(chan checkResult, len(dependencies))
		for _, dependency := range dependencies {
			dependency := dependency
			go func() {
				results <- checkDependency(ctx, dependency)
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
				handleIncompleteChecks(ctx, dependencies, dependenciesResult)
				completed = len(dependencies)
			}
		}

		if !ready {
			writeJSON(w, http.StatusServiceUnavailable, response{Status: "error", Dependencies: dependenciesResult})
			return
		}
		writeJSON(w, http.StatusOK, response{Status: "ok", Dependencies: dependenciesResult})
	}
}

func checkDependency(ctx context.Context, dependency Dependency) checkResult {
	if dependency.Checker == nil {
		dependencyHealth.WithLabelValues(dependency.Name).Set(0)
		return checkResult{name: dependency.Name, status: dependencyResult{Status: "error", Backend: dependency.Backend, Error: "check is not configured"}}
	}
	if err := dependency.Checker.Check(ctx); err != nil {
		dependencyHealth.WithLabelValues(dependency.Name).Set(0)
		l.Log.WithFields(logrus.Fields{
			"dependency": dependency.Name,
			"error":      err,
		}).Error("Readiness dependency check failed")
		return checkResult{name: dependency.Name, status: dependencyResult{Status: "error", Backend: dependency.Backend, Error: "dependency check failed"}}
	}
	if err := ctx.Err(); err != nil {
		dependencyHealth.WithLabelValues(dependency.Name).Set(0)
		l.Log.WithFields(logrus.Fields{
			"dependency": dependency.Name,
			"error":      err,
		}).Error("Readiness dependency check failed")
		return checkResult{name: dependency.Name, status: dependencyResult{Status: "error", Backend: dependency.Backend, Error: "dependency check failed"}}
	}
	dependencyHealth.WithLabelValues(dependency.Name).Set(1)
	return checkResult{name: dependency.Name, status: dependencyResult{Status: "ok", Backend: dependency.Backend}}
}

func handleIncompleteChecks(ctx context.Context, dependencies []Dependency, dependenciesResult map[string]dependencyResult) {
	for _, dependency := range dependencies {
		if _, ok := dependenciesResult[dependency.Name]; ok {
			continue
		}
		l.Log.WithFields(logrus.Fields{
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
