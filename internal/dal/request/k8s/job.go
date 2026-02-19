package k8s

// JobCreateRequest 创建Job请求（表单方式）
type JobCreateRequest struct {
	JobName        string `json:"jobName" binding:"required"`
	NameSpace      string `json:"nameSpace" binding:"required"`
	ContainerName  string `json:"containerName" binding:"required"`
	ContainerImage string `json:"containerImage" binding:"required"`
	Command        string `json:"command" binding:"required"`
}

// YAMLJobCreateRequest 创建Job请求（YAML方式）
type YAMLJobCreateRequest struct {
	YAML string `json:"yaml" binding:"required"`
}

// JobListItem Job列表项
type JobListItem struct {
	JobName        string       `json:"jobName"`
	NameSpace      string       `json:"nameSpace"`
	ContainerName  string       `json:"containerName"`
	ContainerImage string       `json:"containerImage"`
	CommandArgs    string       `json:"commandArgs"`
	Labels         string       `json:"labels"`
	StartTime      string       `json:"startTime"`
	EndTime        string       `json:"endTime"`
	PodsStatuses   string       `json:"podsStatuses"`
	Resources      ResourceInfo `json:"resources"`
}

// JobDetail Job详情
type JobDetail struct {
	JobName        string `json:"jobName"`
	NameSpace      string `json:"nameSpace"`
	ContainerName  string `json:"containerName"`
	ContainerImage string `json:"containerImage"`
	CommandArgs    string `json:"commandArgs"`
	Labels         string `json:"labels"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
	PodsStatuses   string `json:"podsStatuses"`
	Age            int64  `json:"age"`
}
