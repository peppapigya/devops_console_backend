package k8s

import (
	"time"
)

// ChaosExperimentCreateRequest 创建混沌实验请求
type ChaosExperimentCreateRequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	FaultType   string            `json:"faultType" binding:"required"` // PodChaos, NetworkChaos, IOChaos, StressChaos
	Spec        interface{}       `json:"spec" binding:"required"`      // 具体的故障配置
	Duration    string            `json:"duration"`                     // 持续时间,例如 "5m"
	Schedule    *ScheduleSpec     `json:"schedule"`                     // 调度配置,用于定时实验
}

// ScheduleSpec 调度配置
type ScheduleSpec struct {
	Type     string `json:"type"` // cron, once
	Cron     string `json:"cron"` // cron表达式
	Duration string `json:"duration"`
}

// PodChaosSpec Pod混沌实验配置
type PodChaosSpec struct {
	Action        string       `json:"action"` // pod-kill, pod-failure, container-kill
	ContainerName string       `json:"containerName"`
	GracePeriod   int64        `json:"gracePeriod"` // 删除时的宽限期(秒)
	Mode          string       `json:"mode"`        // one, all, fixed, fixed-percent, random-max-percent
	Value         string       `json:"value"`
	Selector      SelectorSpec `json:"selector"`
	Duration      string       `json:"duration"`
}

// NetworkChaosSpec 网络混沌实验配置
type NetworkChaosSpec struct {
	Action          string         `json:"action"` // delay, loss, duplicate, corrupt, partition
	Delay           *DelaySpec     `json:"delay"`
	Loss            *LossSpec      `json:"loss"`
	Duplicate       *DuplicateSpec `json:"duplicate"`
	Corrupt         *CorruptSpec   `json:"corrupt"`
	Direction       string         `json:"direction"` // to, from, both
	ExternalTargets []string       `json:"externalTargets"`
	Selector        SelectorSpec   `json:"selector"`
	Duration        string         `json:"duration"`
}

// DelaySpec 网络延迟配置
type DelaySpec struct {
	Latency     string `json:"latency"`     // 例如 "10ms"
	Jitter      string `json:"jitter"`      // 例如 "5ms"
	Correlation string `json:"correlation"` // 例如 "25"
}

// LossSpec 网络丢包配置
type LossSpec struct {
	Loss        string `json:"loss"`        // 丢包百分比,例如 "10"
	Correlation string `json:"correlation"` // 相关性
}

// DuplicateSpec 网络重复包配置
type DuplicateSpec struct {
	Duplicate   string `json:"duplicate"` // 重复百分比
	Correlation string `json:"correlation"`
}

// CorruptSpec 网络包损坏配置
type CorruptSpec struct {
	Corrupt     string `json:"corrupt"` // 损坏百分比
	Correlation string `json:"correlation"`
}

// IOChaosSpec IO混沌实验配置
type IOChaosSpec struct {
	Action   string       `json:"action"`  // latency, fault
	Percent  int          `json:"percent"` // 影响的百分比
	Path     string       `json:"path"`    // 目标路径
	Methods  []string     `json:"methods"` // 读写方法: read, write, all
	Selector SelectorSpec `json:"selector"`
	Duration string       `json:"duration"`
	Faults   []IOFault    `json:"faults"`
}

// IOFault IO故障配置
type IOFault struct {
	Errno  int    `json:"errno"`  // 错误码
	How    string `json:"how"`    // 错误类型
	Len    int    `json:"len"`    // 长度
	Offset int    `json:"offset"` // 偏移量
}

// StressChaosSpec 压力混沌实验配置
type StressChaosSpec struct {
	CPU      *CPUStressor    `json:"cpu"`
	Memory   *MemoryStressor `json:"memory"`
	Selector SelectorSpec    `json:"selector"`
	Duration string          `json:"duration"`
	Workers  int             `json:"workers"` // worker数量
}

// CPUStressor CPU压力配置
type CPUStressor struct {
	Load    int `json:"load"`    // CPU负载(百分比)
	Workers int `json:"workers"` // worker数量
}

// MemoryStressor 内存压力配置
type MemoryStressor struct {
	Size     string `json:"size"`     // 内存大小,例如 "256MB"
	Workers  int    `json:"workers"`  // worker数量
	Resident bool   `json:"resident"` // 是否驻留
}

// SelectorSpec 选择器配置
type SelectorSpec struct {
	Namespaces          []string            `json:"namespaces"`
	LabelSelectors      map[string]string   `json:"labelSelectors"`
	AnnotationSelectors map[string]string   `json:"annotationSelectors"`
	FieldSelectors      map[string]string   `json:"fieldSelectors"`
	PodSelector         []string            `json:"podSelector"`
	Pods                map[string][]string `json:"pods"` // namespace -> pod names
}

// ChaosExperimentListRequest 查询混沌实验列表请求
type ChaosExperimentListRequest struct {
	Namespace string `form:"namespace"`
	Type      string `form:"type"`   // 按故障类型过滤
	Status    string `form:"status"` // 按状态过滤
}

// ChaosExperimentListItem 混沌实验列表项
type ChaosExperimentListItem struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	FaultType   string            `json:"faultType"`
	Status      string            `json:"status"`
	Phase       string            `json:"phase"`
	CreatedAt   time.Time         `json:"createdAt"`
	StartedAt   *time.Time        `json:"startedAt"`
	FinishedAt  *time.Time        `json:"finishedAt"`
	Duration    string            `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ChaosExperimentDetail 混沌实验详情
type ChaosExperimentDetail struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	FaultType   string            `json:"faultType"`
	Status      string            `json:"status"`
	Phase       string            `json:"phase"`
	CreatedAt   time.Time         `json:"createdAt"`
	StartedAt   *time.Time        `json:"startedAt"`
	FinishedAt  *time.Time        `json:"finishedAt"`
	Duration    string            `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Spec        interface{}       `json:"spec"`    // 具体的故障配置
	Events      []EventInfo       `json:"events"`  // 相关事件
	Targets     []TargetInfo      `json:"targets"` // 影响的目标
}

// EventInfo 事件信息
type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// TargetInfo 目标信息
type TargetInfo struct {
	PodName   string `json:"podName"`
	Namespace string `json:"namespace"`
	NodeName  string `json:"nodeName"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// ChaosExperimentUpdateRequest 更新混沌实验请求
type ChaosExperimentUpdateRequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// ChaosExperimentDeleteRequest 删除混沌实验请求
type ChaosExperimentDeleteRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// ChaosExperimentPauseRequest 暂停混沌实验请求
type ChaosExperimentPauseRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// ChaosExperimentResumeRequest 恢复混沌实验请求
type ChaosExperimentResumeRequest struct {
	Name      string `json:"name" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
}

// ChaosNode 演练节点信息
type ChaosNode struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Status string            `json:"status"` // Ready / NotReady
}

// PrepareEvictionRequest 准备演练节点驱逐请求
type PrepareEvictionRequest struct {
	NodeName            string `json:"nodeName" binding:"required"`            // 演练节点名称
	Namespace           string `json:"namespace" binding:"required"`           // 实验 CR 命名空间
	DeploymentNamespace string `json:"deploymentNamespace" binding:"required"` // Deployment 所在命名空间
	DeploymentName      string `json:"deploymentName" binding:"required"`      // 目标 Deployment 名
}

// CleanupEvictionRequest 清理演练环境请求
type CleanupEvictionRequest struct {
	NodeName            string `json:"nodeName" binding:"required"`
	Namespace           string `json:"namespace" binding:"required"`
	DeploymentNamespace string `json:"deploymentNamespace" binding:"required"`
	DeploymentName      string `json:"deploymentName" binding:"required"`
}
