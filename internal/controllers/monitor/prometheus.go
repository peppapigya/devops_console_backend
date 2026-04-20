package monitor

import "github.com/prometheus/client_golang/prometheus"

// 定义 prometheus 所有的变量

var (
	HttpRequestsTotal    *prometheus.CounterVec
	HttpDuration         *prometheus.HistogramVec
	RepairSessionsTotal  *prometheus.CounterVec
	RepairActionsTotal   *prometheus.CounterVec
	RepairActionDuration *prometheus.HistogramVec
	RepairApprovalsTotal *prometheus.CounterVec
)

func InitPrometheus() {
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "peppapig",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "HTTP URL 请求总数",
		},
		[]string{"method", "path", "status"},
	)

	HttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "peppapig",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP 请求延迟",
		},
		[]string{"path"},
	)

	RepairSessionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "peppapig",
			Subsystem: "repair",
			Name:      "sessions_total",
			Help:      "Total number of repair sessions by status",
		},
		[]string{"status"},
	)

	RepairActionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "peppapig",
			Subsystem: "repair",
			Name:      "actions_total",
			Help:      "Total number of repair actions by tool and status",
		},
		[]string{"tool", "status", "risk_level"},
	)

	RepairActionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "peppapig",
			Subsystem: "repair",
			Name:      "action_duration_seconds",
			Help:      "Repair action execution duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"tool", "status"},
	)

	RepairApprovalsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "peppapig",
			Subsystem: "repair",
			Name:      "approvals_total",
			Help:      "Total number of repair approval decisions",
		},
		[]string{"decision"},
	)

	prometheus.MustRegister(HttpRequestsTotal, HttpDuration, RepairSessionsTotal, RepairActionsTotal, RepairActionDuration, RepairApprovalsTotal)
}
