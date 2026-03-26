package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type K8sExecutor struct{}

func NewK8sExecutor() *K8sExecutor {
	return &K8sExecutor{}
}

func (e *K8sExecutor) GetType() string {
	return "kubernetes"
}

func (e *K8sExecutor) Validate(config map[string]interface{}) error {
	action := getString(config, "action", "")
	if action == "" {
		return fmt.Errorf("操作类型不能为空")
	}
	return nil
}

func (e *K8sExecutor) Execute(ctx context.Context, execCtx *TaskExecutionContext) *ExecutionResult {
	startTime := time.Now()

	k8sInstanceID := getUint64(execCtx.Config, "k8s_instance_id", 0)
	namespace := getString(execCtx.Config, "namespace", "default")
	action := getString(execCtx.Config, "action", "apply")
	yaml := getString(execCtx.Config, "yaml", "")

	execCtx.Logger.Log("info", fmt.Sprintf("开始执行K8s操作，集群ID: %d", k8sInstanceID))
	execCtx.Logger.Log("info", fmt.Sprintf("命名空间: %s, 操作: %s", namespace, action))

	kubeconfig, err := getKubeconfig(k8sInstanceID)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("获取K8s配置失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}

	var cmdArgs []string
	switch action {
	case "apply":
		cmdArgs = []string{"apply", "-f", "-", "-n", namespace}
	case "delete":
		cmdArgs = []string{"delete", "-f", "-", "-n", namespace}
	case "get":
		cmdArgs = []string{"get", "all", "-n", namespace}
	default:
		cmdArgs = []string{action, "-n", namespace}
	}

	cmd := exec.CommandContext(ctx, "kubectl", cmdArgs...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", kubeconfig))

	if yaml != "" {
		cmd.Stdin = bytes.NewBufferString(yaml)
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime).Milliseconds()

	execCtx.Logger.Log("info", "kubectl输出: "+string(output))

	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("K8s操作失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			Output:   string(output),
			ErrorMsg: err.Error(),
			Duration: duration,
		}
	}

	execCtx.Logger.Log("info", fmt.Sprintf("K8s操作成功，耗时: %dms", duration))

	return &ExecutionResult{
		Success:    true,
		Output:     string(output),
		Duration:   duration,
		OutputVars: map[string]interface{}{},
	}
}

func getKubeconfig(k8sInstanceID uint64) (string, error) {
	return "/root/.kube/config", nil
}
