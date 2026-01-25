package indices

import (
	"bytes"
	"context"
	"devops-console-backend/internal/common"
	req "devops-console-backend/internal/dal/request"
	indices3 "devops-console-backend/internal/dal/request/indices"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func Indexcreate(r *gin.Context) {
	indexcreate(r)
}

func Catindices(r *gin.Context) {
	catindices(r)
}

func Deleteindices(r *gin.Context) {
	deleteindices(r)
}

func Updateindices(r *gin.Context) {
	updateindices(r)
}

// deleteindices 索引删除
func deleteindices(r *gin.Context) {
	var indices indices3.IndexDeleteRequest
	returnData := common.NewReturnData()
	// 尝试从查询参数获取instance_id
	instanceID := r.Query("instance_id")
	if instanceID == "" {
		// 如果查询参数没有，尝试从JSON请求体获取
		if err := r.ShouldBindJSON(&indices); err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数绑定失败")
			returnData.Status = 400
			returnData.Message = "缺少参数"
			r.JSON(400, returnData)
			return
		}
	} else {
		// 从查询参数解析instance_id
		var id int
		_, err := fmt.Sscanf(instanceID, "%d", &id)
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数格式错误")
			returnData.Status = 400
			returnData.Message = "参数格式错误"
			r.JSON(400, returnData)
			return
		}
		indices.InstanceID = uint(id)
	}

	client, exists := configs.GetEsClient(indices.InstanceID)
	if !exists || client == nil {
		returnData.Status = 404
		returnData.Message = "该集群不存在"
		r.JSON(404, returnData)
		return
	}
	for _, indexName := range indices.IndexNames {
		res, err := client.Indices.Delete(
			[]string{indexName},
			client.Indices.Delete.WithContext(context.Background()),
		)
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error(), "instance_id": indices.InstanceID}, "Elasticsearch 节点信息请求失败")
			returnData.Status = 500
			returnData.Message = "连接失败"
			r.JSON(500, returnData)
			return
		}

		defer res.Body.Close()

		// 检查响应是否有错误
		if res.IsError() {
			logs.Error(map[string]interface{}{"status": res.Status(), "instance_id": indices.InstanceID}, "Elasticsearch 请求失败")
			returnData.Status = 500
			returnData.Message = "请求失败"
			r.JSON(500, returnData)
			return
		}

		// 成功，返回 ES 返回的原始字符串
		returnData.Status = 200
		returnData.Message = "Elasticsearch 索引删除成功"
		returnData.Data = map[string]interface{}{"raw_response": res.Body}
		r.JSON(200, returnData)

	}

}

// catindices 索引查询
func catindices(r *gin.Context) {
	var connReq req.ConnectionTestRequest
	returnData := common.NewReturnData()
	// 尝试从查询参数获取instance_id
	instanceID := r.Query("instance_id")
	if instanceID == "" {
		// 如果查询参数没有，尝试从JSON请求体获取
		if err := r.ShouldBindJSON(&connReq); err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数绑定失败")
			returnData.Status = 400
			returnData.Message = "缺少参数"
			r.JSON(400, returnData)
			return
		}
	} else {
		// 从查询参数解析instance_id
		var id int
		_, err := fmt.Sscanf(instanceID, "%d", &id)
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数格式错误")
			returnData.Status = 400
			returnData.Message = "参数格式错误"
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

	// 使用 _cat/indices API 获取所有索引
	res, err := client.Cat.Indices(
		client.Cat.Indices.WithContext(context.Background()),
		client.Cat.Indices.WithFormat("json"),
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

	var esResponse []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "解析 Elasticsearch 响应失败")
		returnData.Status = 500
		returnData.Message = "解析响应失败"
		r.JSON(500, returnData)
		return
	}

	// 成功，返回 ES 返回的原始字符串
	returnData.Status = 200
	returnData.Message = "Elasticsearch 索引列表获取成功"
	returnData.Data = map[string]interface{}{"raw_response": esResponse}
	r.JSON(200, returnData)

}

// indexupdate 索引更新
func updateindices(r *gin.Context) {
	var indices indices3.IndexUpdateRequest
	returnData := common.NewReturnData()
	// 尝试从查询参数获取instance_id
	instanceID := r.Query("instance_id")
	if instanceID == "" {
		// 如果查询参数没有，尝试从JSON请求体获取
		if err := r.ShouldBindJSON(&indices); err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数绑定失败")
			returnData.Status = 400
			returnData.Message = "缺少参数"
			r.JSON(400, returnData)
			return
		}
	} else {
		// 从查询参数解析instance_id
		var id int
		_, err := fmt.Sscanf(instanceID, "%d", &id)
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数格式错误")
			returnData.Status = 400
			returnData.Message = "参数格式错误"
			r.JSON(400, returnData)
			return
		}
		indices.InstanceID = uint(id)
	}

	client, exists := configs.GetEsClient(indices.InstanceID)
	if !exists || client == nil {
		returnData.Status = 404
		returnData.Message = "该集群不存在"
		r.JSON(404, returnData)
		return
	}

	body := map[string]map[string]interface{}{"index": {}}
	if indices.NumberOfReplicas != nil {
		body["index"]["number_of_replicas"] = *indices.NumberOfReplicas
	}
	if indices.RefreshInterval != nil {
		body["index"]["refresh_interval"] = *indices.RefreshInterval
	}
	b, _ := json.Marshal(body)
	res, err := client.Indices.PutSettings(
		bytes.NewReader(b),
		client.Indices.PutSettings.WithIndex(indices.IndexName),
		client.Indices.PutSettings.WithContext(context.Background()),
	)
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error(), "instance_id": indices.InstanceID}, "Elasticsearch 节点信息请求失败")
		returnData.Status = 500
		returnData.Message = "连接失败"
		r.JSON(500, returnData)
		return
	}

	defer res.Body.Close()

	// 检查响应是否有错误
	if res.IsError() {
		logs.Error(map[string]interface{}{"status": res.Status(), "instance_id": indices.InstanceID}, "Elasticsearch 请求失败")
		returnData.Status = 400
		returnData.Message = "索引更新失败"
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
	returnData.Message = "Elasticsearch 索引更新成功"
	returnData.Data = map[string]interface{}{"raw_response": esResponse}
	r.JSON(200, returnData)

}

// indexcreate 索引创建
func indexcreate(r *gin.Context) {
	var indices indices3.IndexRequest
	returnData := common.NewReturnData()
	// 尝试从查询参数获取instance_id
	instanceID := r.Query("instance_id")
	if instanceID == "" {
		// 如果查询参数没有，尝试从JSON请求体获取
		if err := r.ShouldBindJSON(&indices); err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数绑定失败")
			returnData.Status = 400
			returnData.Message = "缺少参数"
			r.JSON(400, returnData)
			return
		}
	} else {
		// 从查询参数解析instance_id
		var id int
		_, err := fmt.Sscanf(instanceID, "%d", &id)
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error()}, "请求参数格式错误")
			returnData.Status = 400
			returnData.Message = "参数格式错误"
			r.JSON(400, returnData)
			return
		}
		indices.InstanceID = uint(id)
	}

	client, exists := configs.GetEsClient(indices.InstanceID)
	if !exists || client == nil {
		returnData.Status = 404
		returnData.Message = "该集群不存在"
		r.JSON(404, returnData)
		return
	}

	indexSettings := generateIndexSettings(indices)
	reqBody := bytes.NewBufferString(indexSettings)

	res, err := client.Indices.Create(indices.IndexName, client.Indices.Create.WithBody(reqBody))

	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error(), "instance_id": indices.InstanceID}, "Elasticsearch 节点信息请求失败")
		returnData.Status = 500
		returnData.Message = "连接失败"
		r.JSON(500, returnData)
		return
	}

	defer res.Body.Close()

	// 检查响应是否有错误
	if res.IsError() {
		logs.Error(map[string]interface{}{"status": res.Status(), "instance_id": indices.InstanceID}, "Elasticsearch 请求失败")
		returnData.Status = 400
		returnData.Message = "索引创建失败"
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
	returnData.Message = "Elasticsearch 索引创建成功"
	returnData.Data = map[string]interface{}{"raw_response": esResponse}
	r.JSON(200, returnData)

}

func generateIndexSettings(indices indices3.IndexRequest) string {
	if indices.NumberOfReplicas == 0 {
		indices.NumberOfReplicas = 1
	}
	if indices.RefreshInterval == "" {
		indices.RefreshInterval = "1s"
	}

	// 可选的 analysis 片段
	analysisPart := ""
	if len(indices.Analysis.Tokenizers) != 0 || len(indices.Analysis.Analyzers) != 0 {
		analysisJSON, err := json.Marshal(indices.Analysis)
		if err != nil {
			fmt.Println("Error marshalling analysis:", err)
			return ""
		}
		// 注意这里多了 `"analysis":`
		analysisPart = fmt.Sprintf(`,                                                                                  
        "analysis": %s`, analysisJSON)
	}

	// 可选的 alias 片段
	aliasPart := ""
	if indices.Alias != "" {
		aliasPart = fmt.Sprintf(`,                                                                                     
    "aliases": {                                                                                                         
      "%s": {}                                                                                                           
    }`, indices.Alias)
	}

	// 一次性拼完整 JSON，只保留一个 settings
	indexSettings := fmt.Sprintf(`{                                                                                    
    "settings": {                                                                                                        
      "number_of_shards": %d,                                                                                            
      "number_of_replicas": %d,                                                                                          
      "refresh_interval": "%s"%s                                                                                         
    },                                                                                                                   
    "mappings": {                                                                                                        
      "properties": {                                                                                                    
        "title": { "type": "text" }                                                                                      
      }                                                                                                                  
    }%s                                                                                                                  
  }`,
		indices.NumberOfShards,
		indices.NumberOfReplicas,
		indices.RefreshInterval,
		analysisPart,
		aliasPart,
	)

	return indexSettings
}
