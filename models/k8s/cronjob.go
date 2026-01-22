package k8s

// CronJobCreateRequest 创建CronJob请求
type CronJobCreateRequest struct {
	Name          string   `json:"name" binding:"required"`
	Namespace     string   `json:"namespace" binding:"required"`
	ContainerName string   `json:"containername" binding:"required"`
	Image         string   `json:"image" binding:"required"`
	Command       []string `json:"command"`
	Schedule      string   `json:"schedule" binding:"required"`
}

// CronJobUpdateRequest 更新CronJob请求
type CronJobUpdateRequest struct {
	Name      string  `json:"name" binding:"required"`
	Namespace string  `json:"namespace" binding:"required"`
	Schedule  *string `json:"schedule"`
	Image     *string `json:"image"`
}

// CronJobDeleteRequest 删除CronJob请求
type CronJobDeleteRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// CronJobListItem CronJob列表项
type CronJobListItem struct {
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	ContainerName string   `json:"containername"`
	Image         string   `json:"image"`
	Command       []string `json:"command"`
	Schedule      string   `json:"schedule"`
	Status        string   `json:"status"`
	Age           int64    `json:"age"`
}

// CronJobDetail CronJob详情
type CronJobDetail struct {
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	ContainerName string   `json:"containername"`
	Image         string   `json:"image"`
	Command       []string `json:"command"`
	Schedule      string   `json:"schedule"`
	Status        string   `json:"status"`
	Age           int64    `json:"age"`
	Labels        map[string]string `json:"labels"`
}