package websocket

import (
	"devops-console-backend/pkg/configs"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PodLogHandler WebSocket处理器
type PodLogHandler struct {
	clients map[*websocket.Conn]bool
}

// NewPodLogHandler 创建新的Pod日志处理器
func NewPodLogHandler() *PodLogHandler {
	return &PodLogHandler{
		clients: make(map[*websocket.Conn]bool),
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *PodLogHandler) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源，生产环境应该更严格
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "升级WebSocket失败"})
		return
	}

	h.clients[conn] = true
	defer delete(h.clients, conn)

	// 获取参数
	namespace := c.Query("namespace")
	podName := c.Param("podname")
	container := c.Query("container")
	instanceIdStr := c.Query("instance_id")

	instanceID := uint(1) // 默认值
	if instanceIdStr != "" {
		if id, err := strconv.ParseInt(instanceIdStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		sendError(conn, "K8s客户端未初始化")
		return
	}

	// 获取Pod日志请求
	tailLines := int64(100)
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  container,
		Follow:     true,
		TailLines:  &tailLines,
		Timestamps: true,
		Previous:   false,
	})

	logStream, err := req.Stream(c.Request.Context())
	if err != nil {
		sendError(conn, "获取日志流失败: "+err.Error())
		return
	}
	defer logStream.Close()

	// 读取日志内容
	buf := make([]byte, 1024)
	for {
		n, err := logStream.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if websocket.IsUnexpectedCloseError(err) {
				break
			}
			sendError(conn, "读取日志失败: "+err.Error())
			break
		}

		if n > 0 {
			// 发送日志内容
			message := map[string]interface{}{
				"type":    "log",
				"content": string(buf[:n]),
				"time":    time.Now().Unix(),
			}
			if err := conn.WriteJSON(message); err != nil {
				break
			}
		}
	}
}

// sendError 发送错误消息
func sendError(conn *websocket.Conn, message string) {
	errorMsg := map[string]interface{}{
		"type":  "error",
		"error": message,
		"time":  time.Now().Unix(),
	}
	conn.WriteJSON(errorMsg)
}

// getK8sClient 获取K8s客户端
func getK8sClient(instanceIdStr string) (*kubernetes.Clientset, error) {
	if instanceIdStr == "" {
		return nil, fmt.Errorf("instance_id不能为空")
	}

	instanceID64, _ := strconv.ParseInt(instanceIdStr, 10, 32)
	instanceID := uint(instanceID64)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return nil, fmt.Errorf("instance_id %d 对应的K8s客户端未初始化", instanceID)
	}

	return client, nil
}
