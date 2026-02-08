package k8s

// EventListItem Event列表项
type EventListItem struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Type           string `json:"type"` // Normal, Warning
	Reason         string `json:"reason"`
	Message        string `json:"message"`
	InvolvedObject string `json:"involvedObject"`
	InvolvedKind   string `json:"involvedKind"`
	Source         string `json:"source"`
	Count          int32  `json:"count"`
	FirstTimestamp int64  `json:"firstTimestamp"`
	LastTimestamp  int64  `json:"lastTimestamp"`
}
