package instance

import (
	"devops-console-backend/utils/logs"
	"github.com/gin-gonic/gin"
)

// Add 添加Elasticsearch实例
// @Summary 添加Elasticsearch实例
// @Description 添加新的Elasticsearch实例配置
// @Tags instance
// @Accept json
// @Produce json
// @Param instance body models.Instance true "实例信息"
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/add [post]
func Add(r *gin.Context) {
	logs.Debug(nil, "添加集群")
	handleInstanceOperation(r, true)
}
