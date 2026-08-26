// Package securitylog provides structured security event logging for
// SEC-MON-REQ-1 compliance (Events of Interest).
//
// Every security log entry is a JSON object emitted via logrus with a
// "security_event": true marker so that log aggregation pipelines can
// filter on it. Each entry carries the five required fields: action,
// resource_type, resource_id, outcome, and principal.
package securitylog

import (
	"github.com/sirupsen/logrus"
)

// Principal identifies who performed the action.
type Principal struct {
	UserID  string `json:"user_id,omitempty"`
	OrgID   string `json:"org_id,omitempty"`
	Type    string `json:"type"` // "User", "ServiceAccount", "System", "anonymous"
}

// Event holds the data for a single security log entry.
type Event struct {
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Outcome      string    `json:"outcome"` // "success" or "failure"
	Principal    Principal `json:"principal"`
	Reason       string    `json:"reason,omitempty"`
}

// Log emits a security event at the appropriate log level.
// Success events use Info; failure events use Warn.
func Log(logger *logrus.Logger, e Event) {
	fields := logrus.Fields{
		"security_event": true,
		"action":         e.Action,
		"resource_type":  e.ResourceType,
		"resource_id":    e.ResourceID,
		"outcome":        e.Outcome,
		"principal": logrus.Fields{
			"user_id": e.Principal.UserID,
			"org_id":  e.Principal.OrgID,
			"type":    e.Principal.Type,
		},
	}
	if e.Reason != "" {
		fields["reason"] = e.Reason
	}

	entry := logger.WithFields(fields)
	if e.Outcome == "failure" {
		entry.Warn("security event")
	} else {
		entry.Info("security event")
	}
}

// LogStartup records a process lifecycle STARTUP event (EOI-5).
func LogStartup(logger *logrus.Logger) {
	Log(logger, Event{
		Action:       "STARTUP",
		ResourceType: "process",
		ResourceID:   "insights-ingress-go",
		Outcome:      "success",
		Principal:    Principal{Type: "System"},
	})
}

// LogShutdown records a process lifecycle SHUTDOWN event (EOI-5).
// outcome should be "success" for graceful shutdown or "failure" for crash.
func LogShutdown(logger *logrus.Logger, outcome, reason string) {
	Log(logger, Event{
		Action:       "SHUTDOWN",
		ResourceType: "process",
		ResourceID:   "insights-ingress-go",
		Outcome:      outcome,
		Principal:    Principal{Type: "System"},
		Reason:       reason,
	})
}

// LogAuthFailure records an authentication failure event (EOI-7).
func LogAuthFailure(logger *logrus.Logger, reason, resourceID string) {
	Log(logger, Event{
		Action:       "AUTH_FAILURE",
		ResourceType: "request",
		ResourceID:   resourceID,
		Outcome:      "failure",
		Principal:    Principal{Type: "anonymous"},
		Reason:       reason,
	})
}
