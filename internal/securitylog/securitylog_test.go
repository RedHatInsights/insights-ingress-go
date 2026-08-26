package securitylog

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
)

// newTestLogger returns a logrus.Logger that writes JSON to the given buffer.
func newTestLogger(buf *bytes.Buffer) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	return logger
}

// parseLine decodes a single JSON log line into a map.
func parseLine(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("failed to parse log line: %v\nraw: %s", err, buf.String())
	}
	return m
}

func TestLog_SuccessEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	Log(logger, Event{
		Action:       "CREATE",
		ResourceType: "payload",
		ResourceID:   "req-123",
		Outcome:      "success",
		Principal:    Principal{UserID: "user-1", OrgID: "org-1", Type: "User"},
	})

	m := parseLine(t, buf)

	if m["security_event"] != true {
		t.Error("expected security_event=true")
	}
	if m["action"] != "CREATE" {
		t.Errorf("expected action=CREATE, got %v", m["action"])
	}
	if m["resource_type"] != "payload" {
		t.Errorf("expected resource_type=payload, got %v", m["resource_type"])
	}
	if m["resource_id"] != "req-123" {
		t.Errorf("expected resource_id=req-123, got %v", m["resource_id"])
	}
	if m["outcome"] != "success" {
		t.Errorf("expected outcome=success, got %v", m["outcome"])
	}
	if m["level"] != "info" {
		t.Errorf("expected level=info for success, got %v", m["level"])
	}

	principal, ok := m["principal"].(map[string]interface{})
	if !ok {
		t.Fatal("expected principal to be a map")
	}
	if principal["user_id"] != "user-1" {
		t.Errorf("expected principal.user_id=user-1, got %v", principal["user_id"])
	}
	if principal["org_id"] != "org-1" {
		t.Errorf("expected principal.org_id=org-1, got %v", principal["org_id"])
	}
	if principal["type"] != "User" {
		t.Errorf("expected principal.type=User, got %v", principal["type"])
	}
}

func TestLog_FailureEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	Log(logger, Event{
		Action:       "CREATE",
		ResourceType: "payload",
		ResourceID:   "req-456",
		Outcome:      "failure",
		Principal:    Principal{OrgID: "org-2", Type: "User"},
		Reason:       "staging error",
	})

	m := parseLine(t, buf)

	if m["outcome"] != "failure" {
		t.Errorf("expected outcome=failure, got %v", m["outcome"])
	}
	if m["level"] != "warning" {
		t.Errorf("expected level=warning for failure, got %v", m["level"])
	}
	if m["reason"] != "staging error" {
		t.Errorf("expected reason='staging error', got %v", m["reason"])
	}
}

func TestLog_ReasonOmittedWhenEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	Log(logger, Event{
		Action:       "CREATE",
		ResourceType: "payload",
		ResourceID:   "req-789",
		Outcome:      "success",
		Principal:    Principal{Type: "User"},
	})

	m := parseLine(t, buf)

	if _, exists := m["reason"]; exists {
		t.Error("reason field should be omitted when empty")
	}
}

func TestLogStartup(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	LogStartup(logger)

	m := parseLine(t, buf)

	if m["action"] != "STARTUP" {
		t.Errorf("expected action=STARTUP, got %v", m["action"])
	}
	if m["resource_type"] != "process" {
		t.Errorf("expected resource_type=process, got %v", m["resource_type"])
	}
	if m["resource_id"] != "insights-ingress-go" {
		t.Errorf("expected resource_id=insights-ingress-go, got %v", m["resource_id"])
	}
	if m["outcome"] != "success" {
		t.Errorf("expected outcome=success, got %v", m["outcome"])
	}

	principal, ok := m["principal"].(map[string]interface{})
	if !ok {
		t.Fatal("expected principal to be a map")
	}
	if principal["type"] != "System" {
		t.Errorf("expected principal.type=System, got %v", principal["type"])
	}
}

func TestLogShutdown_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	LogShutdown(logger, "success", "")

	m := parseLine(t, buf)

	if m["action"] != "SHUTDOWN" {
		t.Errorf("expected action=SHUTDOWN, got %v", m["action"])
	}
	if m["outcome"] != "success" {
		t.Errorf("expected outcome=success, got %v", m["outcome"])
	}
	if m["level"] != "info" {
		t.Errorf("expected level=info for successful shutdown, got %v", m["level"])
	}
	if _, exists := m["reason"]; exists {
		t.Error("reason should be omitted for successful shutdown")
	}
}

func TestLogShutdown_Failure(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	LogShutdown(logger, "failure", "unexpected error")

	m := parseLine(t, buf)

	if m["outcome"] != "failure" {
		t.Errorf("expected outcome=failure, got %v", m["outcome"])
	}
	if m["level"] != "warning" {
		t.Errorf("expected level=warning for failed shutdown, got %v", m["level"])
	}
	if m["reason"] != "unexpected error" {
		t.Errorf("expected reason='unexpected error', got %v", m["reason"])
	}
}

func TestLogAuthFailure(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	LogAuthFailure(logger, "invalid identity header", "req-auth-1")

	m := parseLine(t, buf)

	if m["action"] != "AUTH_FAILURE" {
		t.Errorf("expected action=AUTH_FAILURE, got %v", m["action"])
	}
	if m["resource_type"] != "request" {
		t.Errorf("expected resource_type=request, got %v", m["resource_type"])
	}
	if m["resource_id"] != "req-auth-1" {
		t.Errorf("expected resource_id=req-auth-1, got %v", m["resource_id"])
	}
	if m["outcome"] != "failure" {
		t.Errorf("expected outcome=failure, got %v", m["outcome"])
	}
	if m["reason"] != "invalid identity header" {
		t.Errorf("expected reason='invalid identity header', got %v", m["reason"])
	}

	principal, ok := m["principal"].(map[string]interface{})
	if !ok {
		t.Fatal("expected principal to be a map")
	}
	if principal["type"] != "anonymous" {
		t.Errorf("expected principal.type=anonymous, got %v", principal["type"])
	}
}

func TestLog_AllRequiredFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	Log(logger, Event{
		Action:       "DELETE",
		ResourceType: "payload",
		ResourceID:   "req-del-1",
		Outcome:      "success",
		Principal:    Principal{UserID: "u1", OrgID: "o1", Type: "ServiceAccount"},
	})

	m := parseLine(t, buf)

	requiredFields := []string{"security_event", "action", "resource_type", "resource_id", "outcome", "principal"}
	for _, field := range requiredFields {
		if _, exists := m[field]; !exists {
			t.Errorf("required field %q missing from security event log", field)
		}
	}
}
