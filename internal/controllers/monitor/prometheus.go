package monitor

import "github.com/prometheus/client_golang/prometheus"

// 定义 prometheus 所有的变量

var (
	HttpRequestsTotal *prometheus.CounterVec
	HttpDuration      *prometheus.HistogramVec
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

	prometheus.MustRegister(HttpRequestsTotal, HttpDuration)
}
