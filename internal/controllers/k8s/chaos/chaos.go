package chaos

import (
	"devops-console-backend/internal/dal/request/k8s"
	chaos_service "devops-console-backend/internal/services/k8s/chaos"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChaosController struct {
	service *chaos_service.ChaosService
}

func NewChaosController() *ChaosController {
	return &ChaosController{
		service: chaos_service.NewChaosService(),
	}
}

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
