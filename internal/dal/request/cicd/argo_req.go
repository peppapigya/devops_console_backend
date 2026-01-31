package cicd

type Step struct {
	Name    string            `json:"name"`    // 步骤实例名
	TaskKey string            `json:"taskKey"` // 引用组件
	Params  map[string]string `json:"params"`  // 用户填写的参数
	Depends []string          `json:"depends"` // 依赖项,后期去扩展
}

type RunReq struct {
	PipelineID uint   `json:"pipelineId"`
	Steps      []Step `json:"steps"`
}
