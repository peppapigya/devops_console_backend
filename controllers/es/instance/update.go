package instance

import (
	"devops-console-backend/utils/logs"
	"github.com/gin-gonic/gin"
)

// Update 更新Elasticsearch实例
// @Summary 更新Elasticsearch实例
// @Description 更新现有Elasticsearch实例的配置信息
// @Tags instance
// @Accept json
// @Produce json
// @Param instance body models.Instance true "实例信息"
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/update [post]
func Update(r *gin.Context) {
	logs.Debug(nil, "修改集群")
	handleInstanceOperation(r, false)
}
