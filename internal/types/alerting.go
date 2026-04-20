package types

import "time"

type AlertmanagerWebhookPayload struct {
	Version           string              `json:"version"`
	GroupKey          string              `json:"groupKey"`
	Status            string              `json:"status"`
	Receiver          string              `json:"receiver"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotations"`
	ExternalURL       string              `json:"externalURL"`
	Alerts            []AlertmanagerAlert `json:"alerts"`
}

type AlertmanagerAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type NormalizedAlert struct {
	Fingerprint string                 `json:"fingerprint"`
	AlertName   string                 `json:"alert_name"`
	AlertKey    string                 `json:"alert_key"`
	AlertStatus string                 `json:"alert_status"`
	Source      string                 `json:"source"`
	Host        string                 `json:"host"`
	Service     string                 `json:"service"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	LogEvent    LogEvent               `json:"log_event"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type AlertDispatchResult struct {
	Fingerprint string `json:"fingerprint"`
	AlertStatus string `json:"alert_status"`
	Dispatch    string `json:"dispatch"`
	SessionID   string `json:"session_id,omitempty"`
	Message     string `json:"message,omitempty"`
}
