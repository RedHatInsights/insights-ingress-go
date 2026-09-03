package health

import (
	prom "github.com/prometheus/client_golang/prometheus"
	pa "github.com/prometheus/client_golang/prometheus/promauto"
)

var dependencyHealth = pa.NewGaugeVec(prom.GaugeOpts{
	Name: "ingress_dependency_health",
	Help: "Health of an Ingress readiness dependency (1 is healthy, 0 is unhealthy).",
}, []string{"dependency"})
