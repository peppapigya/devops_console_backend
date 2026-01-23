package node

import (
	"context"
	"devops-console-backend/internal/common"
	req "devops-console-backend/internal/models/request"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

// NodeInfo 结构体定义节点信息
type NodeInfo struct {
	Name    string  `json:"name"`
	NodeId  string  `json:"node_id"`
	IP      string  `json:"ip"`
	Address string  `json:"address"`
	Version string  `json:"version"`
	Role    string  `json:"role"`
	JVMInfo JVMInfo `json:"jvm_info"`
	OSInfo  OSInfo  `json:"os_info"`
}

// JVMInfo JVM信息结构体
type JVMInfo struct {
	Version    string     `json:"version"`
	Vendor     string     `json:"vendor"`
	VMName     string     `json:"vm_name"`
	VMVersion  string     `json:"vm_version"`
	StartTime  int64      `json:"start_time"`
	MemoryInfo MemoryInfo `json:"memory_info"`
}

// OSInfo 操作系统信息结构体
type OSInfo struct {
	Name                string `json:"name"`
	Version             string `json:"version"`
	Architecture        string `json:"architecture"`
	AvailableProcessors int    `json:"available_processors"`
}

// MemoryInfo 内存信息结构体
type MemoryInfo struct {
	HeapInitSize string `json:"heap_init_size"`
	HeapMaxSize  string `json:"heap_max_size"`
}

func ClusterNodeInfo(r *gin.Context) {
	NodeInfoResponse(r)
}

// NodeInfoResponse 获取节点详细信息
func NodeInfoResponse(r *gin.Context) {
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

	client, exists := configs.GetEsClient(connReq.InstanceID)
	if !exists || client == nil {
		returnData.Status = 404
		returnData.Message = "该集群不存在"
		r.JSON(404, returnData)
		return
	}
	// 获取节点信息
	res, err := client.Nodes.Info(
		client.Nodes.Info.WithContext(context.Background()),
		client.Nodes.Info.WithNodeID(""),
		client.Nodes.Info.WithMetric("os", "jvm", "process", "settings"),
	)

	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error(), "instance_id": connReq.InstanceID}, "Elasticsearch 节点信息请求失败")
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

	// 解析节点信息
	nodes := make([]NodeInfo, 0)
	if nodesMap, ok := esResponse["nodes"].(map[string]interface{}); ok {
		for nodeId, nodeData := range nodesMap {
			if node, ok := nodeData.(map[string]interface{}); ok {
				// 构建节点信息
				nodeInfo := NodeInfo{
					NodeId: nodeId,
				}

				if name, ok := node["name"].(string); ok {
					nodeInfo.Name = name
				}

				if ip, ok := node["ip"].(string); ok {
					nodeInfo.IP = ip
				}

				if address, ok := node["address"].(string); ok {
					nodeInfo.Address = address
				}

				if version, ok := node["version"].(string); ok {
					nodeInfo.Version = version
				}

				// 提取JVM信息
				if jvm, ok := node["jvm"].(map[string]interface{}); ok {
					nodeInfo.JVMInfo = JVMInfo{
						Version:   getStringValue(jvm, "version"),
						Vendor:    getStringValue(jvm, "vm_vendor"),
						VMName:    getStringValue(jvm, "vm_name"),
						VMVersion: getStringValue(jvm, "vm_version"),
					}
					if startTime, ok := jvm["start_time_in_millis"].(float64); ok {
						nodeInfo.JVMInfo.StartTime = int64(startTime)
					}
					if mem, ok := jvm["mem"].(map[string]interface{}); ok {
						nodeInfo.JVMInfo.MemoryInfo = MemoryInfo{
							HeapInitSize: getStringValue(mem, "heap_init_in_bytes"),
							HeapMaxSize:  getStringValue(mem, "heap_max_in_bytes"),
						}
					}
				}

				// 提取OS信息
				if os, ok := node["os"].(map[string]interface{}); ok {
					nodeInfo.OSInfo = OSInfo{
						Name:                getStringValue(os, "name"),
						Version:             getStringValue(os, "version"),
						Architecture:        getStringValue(os, "arch"),
						AvailableProcessors: getIntValue(os, "available_processors"),
					}
				}

				// 尝试从settings提取角色信息
				if settings, ok := node["settings"].(map[string]interface{}); ok {
					if nodeSettings, ok := settings["node"].(map[string]interface{}); ok {
						roles := []string{}
						if rolesData, ok := nodeSettings["roles"].([]interface{}); ok {
							for _, role := range rolesData {
								if roleStr, ok := role.(string); ok {
									roles = append(roles, roleStr)
								}
							}
						}
						if len(roles) > 0 {
							nodeInfo.Role = fmt.Sprintf("%v", roles)
						}
					}
				}

				nodes = append(nodes, nodeInfo)
			}
		}
	}

	// 成功，返回 ES 返回的原始响应
	returnData.Status = 200
	returnData.Message = "Elasticsearch 节点信息获取成功"
	returnData.Data = map[string]interface{}{"raw_response": esResponse}
	r.JSON(200, returnData)
}

// getStringValue 从map中获取字符串值的辅助函数
func getStringValue(m map[string]interface{}, key string) string {
	if value, ok := m[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// getIntValue 从map中获取整数值的辅助函数
func getIntValue(m map[string]interface{}, key string) int {
	if value, ok := m[key]; ok {
		if num, ok := value.(float64); ok {
			return int(num)
		}
	}
	return 0
}
