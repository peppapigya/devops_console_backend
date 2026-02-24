package request

type CreateCustomMonitorRequest struct {
	TargetType     string `json:"target_type" binding:"required,oneof=pod node"`
	Title          string `json:"title" binding:"required"`
	PromQLTemplate string `json:"promql_template" binding:"required"`
	ChartType      string `json:"chart_type" binding:"required,oneof=line bar"`
	UnitSuffix     string `json:"unit_suffix"`
	ColorTheme     string `json:"color_theme"`
}

type UpdateCustomMonitorRequest struct {
	Title          *string `json:"title"`
	PromQLTemplate *string `json:"promql_template"`
	ChartType      *string `json:"chart_type" binding:"omitempty,oneof=line bar"`
	UnitSuffix     *string `json:"unit_suffix"`
	ColorTheme     *string `json:"color_theme"`
}
