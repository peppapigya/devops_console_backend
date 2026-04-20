package alerting

import (
	"devops-console-backend/internal/types"
	"testing"
	"time"
)

func TestNormalizeAlert(t *testing.T) {
	payload := types.AlertmanagerWebhookPayload{
		Status:   "firing",
		Receiver: "default",
		CommonLabels: map[string]string{
			"service": "order-service",
		},
		CommonAnnotations: map[string]string{
			"description": "default description",
		},
	}
	alert := types.AlertmanagerAlert{
		Status:      "firing",
		Fingerprint: "fp-1",
		StartsAt:    time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
		Labels: map[string]string{
			"alertname": "HighErrorRate",
			"instance":  "10.0.0.8:9090",
			"severity":  "critical",
		},
		Annotations: map[string]string{
			"summary": "5xx spikes detected",
		},
	}

	normalized := normalizeAlert(payload, alert)
	if normalized.Fingerprint != "fp-1" {
		t.Fatalf("unexpected fingerprint: %s", normalized.Fingerprint)
	}
	if normalized.Host != "10.0.0.8" {
		t.Fatalf("unexpected host: %s", normalized.Host)
	}
	if normalized.Service != "order-service" {
		t.Fatalf("unexpected service: %s", normalized.Service)
	}
	if normalized.Level != "CRITICAL" {
		t.Fatalf("unexpected level: %s", normalized.Level)
	}
	if normalized.Message != "5xx spikes detected" {
		t.Fatalf("unexpected message: %s", normalized.Message)
	}
	if normalized.LogEvent.ID == "" || normalized.LogEvent.Source == "" {
		t.Fatalf("normalized log event should be populated: %+v", normalized.LogEvent)
	}
}

func TestBuildRcaLinkMarkdown(t *testing.T) {
	link := buildRcaLinkMarkdown("http://localhost:5173/", "session-1")
	want := "**查看详情**: [打开 RCA 页面](http://localhost:5173/monitor/rca?session_id=session-1)"
	if link != want {
		t.Fatalf("unexpected link: %s", link)
	}
}

func TestNormalizeLevel(t *testing.T) {
	cases := map[string]string{
		"critical": "CRITICAL",
		"warning":  "WARN",
		"resolved": "INFO",
		"":         "ERROR",
	}
	for input, want := range cases {
		got := normalizeLevel(input)
		if got != want {
			t.Fatalf("normalizeLevel(%q) = %q, want %q", input, got, want)
		}
	}
}
