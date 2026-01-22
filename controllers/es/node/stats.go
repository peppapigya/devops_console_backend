package node

import (
	"context"
	"devops-console-backend/common"
	"devops-console-backend/config"
	req "devops-console-backend/models/request"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

// ClusterNodeStats 获取节点统计信息
func ClusterNodeStats(r *gin.Context) {
	var connReq req.ConnectionTestRequest
	returnData := common.NewReturnData()

	// 尝试从查询参数获取instance_id
	instanceID := r.Query("instance_id")
	if instanceID == "" {
		// 如果查询参数没有，尝试从JSON请求体获取
		if err := r.ShouldBindJSON(&connReq); err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数绑定失败")
			returnData.Status = 400
			returnData.Message = "缺少instance_id参数"
			r.JSON(400, returnData)
			return
		}
	} else {
		// 从查询参数解析instance_id
		var id int
		_, err := fmt.Sscanf(instanceID, "%d", &id)
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "instance_id参数格式错误")
			returnData.Status = 400
			returnData.Message = "instance_id参数格式错误"
			r.JSON(400, returnData)
			return
		}
		connReq.InstanceID = uint(id)
	}

	client, exists := config.GetEsClient(connReq.InstanceID)
	if !exists || client == nil {
		returnData.Status = 404
		returnData.Message = "该集群不存在"
		r.JSON(404, returnData)
		return
	}

	// 获取节点统计信息
	res, err := client.Nodes.Stats(
		client.Nodes.Stats.WithContext(context.Background()),
		client.Nodes.Stats.WithNodeID(""),
		client.Nodes.Stats.WithMetric("os", "jvm", "process", "fs", "indices", "thread_pool"),
	)

	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error(), "instance_id": connReq.InstanceID}, "Elasticsearch 节点统计信息请求失败")
		returnData.Status = 500
		returnData.Message = "连接失败"
		r.JSON(500, returnData)
		return
	}

	defer res.Body.Close()

	// 检查响应是否有错误
	if res.IsError() {
		logs.Error(map[string]interface{}{"status": res.Status(), "instance_id": connReq.InstanceID}, "Elasticsearch 请求失败")
		returnData.Status = 500
		returnData.Message = "请求失败"
		r.JSON(500, returnData)
		return
	}

	var esResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "解析 Elasticsearch 响应失败")
		returnData.Status = 500
		returnData.Message = "解析响应失败"
		r.JSON(500, returnData)
		return
	}

	// 成功，返回 ES 返回的原始字符串
	returnData.Status = 200
	returnData.Message = "Elasticsearch 节点统计信息获取成功"
	returnData.Data = map[string]interface{}{"raw_response": esResponse}
	r.JSON(200, returnData)
}
