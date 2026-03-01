package incident

// ==================== 故障管理 ====================

type IncidentPageReq struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=200"`
	BusinessLine string `form:"businessLine"`
	Level        string `form:"level"`  // P1/P2/P3/P4
	Status       string `form:"status"` // pending/processing/done
	Dept         string `form:"dept"`
}

type IncidentCreateReq struct {
	AlertTime    string `json:"alertTime" binding:"required"`
	BusinessLine string `json:"businessLine" binding:"required"`
	Level        string `json:"level" binding:"required,oneof=P1 P2 P3 P4"`
	Frequency    string `json:"frequency"`
	AlertDesc    string `json:"alertDesc" binding:"required"`
	Detail       string `json:"detail"`
	Dept         string `json:"dept" binding:"required"`
	Handler      string `json:"handler" binding:"required"`
	Status       string `json:"status"`
	ResolvedAt   string `json:"resolvedAt"`
}

type IncidentUpdateReq struct {
	AlertTime    string `json:"alertTime"`
	BusinessLine string `json:"businessLine" binding:"required"`
	Level        string `json:"level" binding:"required,oneof=P1 P2 P3 P4"`
	Frequency    string `json:"frequency"`
	AlertDesc    string `json:"alertDesc" binding:"required"`
	Detail       string `json:"detail"`
	Dept         string `json:"dept" binding:"required"`
	Handler      string `json:"handler" binding:"required"`
	Status       string `json:"status"`
	ResolvedAt   string `json:"resolvedAt"`
}

type IncidentStatusReq struct {
	Status string `json:"status" binding:"required,oneof=pending processing done"`
}
