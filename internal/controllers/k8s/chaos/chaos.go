package chaos

import (
	"devops-console-backend/internal/dal/request/k8s"
	chaosService "devops-console-backend/internal/services/k8s/chaos"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChaosController struct {
	service         *chaosService.ChaosService
	evictionService *chaosService.EvictionService
}

func NewChaosController() *ChaosController {
	return &ChaosController{
		service:         chaosService.NewChaosService(),
		evictionService: chaosService.NewEvictionService(),
	}
}

// CreateFault 创建故障
func (c *ChaosController) CreateFault(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	logData := map[string]interface{}{"namespace": namespace}
	logs.Debug(logData, "Create chaos fault")

	var req k8s.ChaosExperimentCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logs.Error(logData, "Bind JSON failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("Invalid request: " + err.Error())
		return
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	if err := c.service.CreateChaosExperiment(ctx, instanceID, &req); err != nil {
		logs.Error(logData, "Create chaos experiment failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("Create chaos experiment failed: " + err.Error())
		return
	}

	logs.Info(logData, "Create chaos experiment success")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("Chaos experiment created successfully")
}

// ListFaults 列出指定命名空间下的所有故障
func (c *ChaosController) ListFaults(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	logData := map[string]interface{}{"namespace": namespace}
	logs.Debug(logData, "List chaos faults")

	var req k8s.ChaosExperimentListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		logs.Error(logData, "Bind query failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("Invalid query: " + err.Error())
		return
	}

	if req.Namespace == "" {
		req.Namespace = namespace
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	list, err := c.service.ListChaosExperiments(ctx, instanceID, &req)
	if err != nil {
		logs.Error(logData, "List chaos experiments failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("List chaos experiments failed: " + err.Error())
		return
	}

	logs.Info(logData, "List chaos experiments success")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "faults", list)
}

// GetFault 获取指定名称的故障详情
func (c *ChaosController) GetFault(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	logData := map[string]interface{}{"namespace": namespace, "name": name}
	logs.Debug(logData, "Get chaos fault")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	detail, err := c.service.GetChaosExperiment(ctx, instanceID, namespace, name)
	if err != nil {
		logs.Error(logData, "Get chaos experiment failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("Get chaos experiment failed: " + err.Error())
		return
	}

	logs.Info(logData, "Get chaos experiment success")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "detail", detail)
}

// DeleteFault 删除指定名称的故障
func (c *ChaosController) DeleteFault(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	logData := map[string]interface{}{"namespace": namespace, "name": name}
	logs.Debug(logData, "Delete chaos fault")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	if err := c.service.DeleteChaosExperiment(ctx, instanceID, namespace, name); err != nil {
		logs.Error(logData, "Delete chaos experiment failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("Delete chaos experiment failed: " + err.Error())
		return
	}

	logs.Info(logData, "Delete chaos experiment success")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("Chaos experiment deleted successfully")
}

// PauseFault 暂停指定名称的故障
func (c *ChaosController) PauseFault(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	logData := map[string]interface{}{"namespace": namespace, "name": name}
	logs.Debug(logData, "Pause chaos fault")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	if err := c.service.PauseChaosExperiment(ctx, instanceID, namespace, name); err != nil {
		logs.Error(logData, "Pause chaos experiment failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("Pause chaos experiment failed: " + err.Error())
		return
	}

	logs.Info(logData, "Pause chaos experiment success")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("Chaos experiment paused successfully")
}

// ResumeFault 恢复指定名称的故障
func (c *ChaosController) ResumeFault(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	logData := map[string]interface{}{"namespace": namespace, "name": name}
	logs.Debug(logData, "Resume chaos fault")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	if err := c.service.ResumeChaosExperiment(ctx, instanceID, namespace, name); err != nil {
		logs.Error(logData, "Resume chaos experiment failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("Resume chaos experiment failed: " + err.Error())
		return
	}

	logs.Info(logData, "Resume chaos experiment success")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("Chaos experiment resumed successfully")
}

// GetChaosNodes 获取所有演练节点（role=chaos-testing）
func (c *ChaosController) GetChaosNodes(ctx *gin.Context) {
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	nodes, err := c.evictionService.GetChaosTestingNodes(ctx, instanceID)
	if err != nil {
		logs.Error(map[string]interface{}{}, "Get chaos testing nodes failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取演练节点失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "nodes", nodes)
}

// PrepareEviction 执行完整的演练节点准备流程（Taint → Patch → Evict → Wait）
func (c *ChaosController) PrepareEviction(ctx *gin.Context) {
	var req k8s.PrepareEvictionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("Invalid request: " + err.Error())
		return
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	logData := map[string]interface{}{
		"node":       req.NodeName,
		"namespace":  req.Namespace,
		"deployment": req.DeploymentName,
	}
	logs.Info(logData, "Prepare eviction start")

	origSpecJSON, err := c.evictionService.PrepareEviction(ctx, instanceID, &req)
	if err != nil {
		logs.Error(logData, "Prepare eviction failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("演练节点准备失败: " + err.Error())
		return
	}

	// 现在我们使用 K8s Annotation 来存储，不需要保存在本地内存中了
	// 返回的 origSpecJSON 在这里可以忽略或记录
	_ = origSpecJSON

	logs.Info(logData, "Prepare eviction success")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("演练节点准备完成，Pod 已迁移到演练节点")
}

// CleanupEviction 清理演练环境（回滚 Deployment → re-Evict → 去 Taint）
func (c *ChaosController) CleanupEviction(ctx *gin.Context) {
	var req k8s.CleanupEvictionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("Invalid request: " + err.Error())
		return
	}

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	logData := map[string]interface{}{
		"node":       req.NodeName,
		"namespace":  req.Namespace,
		"deployment": req.DeploymentName,
	}

	logs.Info(logData, "Cleanup eviction start")

	// 直接传空字符串作为备用 spec，服务内部会优先读取 Deployment Annotation 如果读不到才会报错
	if err := c.evictionService.CleanupEviction(ctx, instanceID, &req, ""); err != nil {
		logs.Error(logData, "Cleanup eviction failed: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("清理演练环境失败: " + err.Error())
		return
	}

	logs.Info(logData, "Cleanup eviction success")
	helper := utils.NewResponseHelper(ctx)
	helper.Success("演练环境清理完成，Pod 已恢复正常调度")
}
