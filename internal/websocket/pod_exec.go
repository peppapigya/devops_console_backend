package websocket

import (
	"devops-console-backend/pkg/configs"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// PodExecHandler WebSocket处理器
type PodExecHandler struct {
	clients map[*websocket.Conn]bool
}

// NewPodExecHandler 创建新的Pod Exec处理器
func NewPodExecHandler() *PodExecHandler {
	return &PodExecHandler{
		clients: make(map[*websocket.Conn]bool),
	}
}

// TerminalMessage 终端消息结构
type TerminalMessage struct {
	Type string `json:"type"` // "stdin", "resize"
	Data string `json:"data"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// TerminalSession 终端会话
type TerminalSession struct {
	conn      *websocket.Conn
	sizeChan  chan remotecommand.TerminalSize
	closeChan chan struct{}
	rows      uint16
	cols      uint16
}

// Read 实现io.Reader接口
func (t *TerminalSession) Read(p []byte) (int, error) {
	var msg TerminalMessage
	err := t.conn.ReadJSON(&msg)
	if err != nil {
		if err == io.EOF || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			return 0, io.EOF
		}
		return 0, err
	}

	switch msg.Type {
	case "stdin":
		copy(p, msg.Data)
		return len(msg.Data), nil
	case "resize":
		if msg.Rows > 0 && msg.Cols > 0 {
			t.rows = msg.Rows
			t.cols = msg.Cols
			t.sizeChan <- remotecommand.TerminalSize{
				Width:  msg.Cols,
				Height: msg.Rows,
			}
		}
	}
	return 0, nil
}

// Write 实现io.Writer接口
func (t *TerminalSession) Write(p []byte) (int, error) {
	msg := map[string]interface{}{
		"type": "stdout",
		"data": string(p),
	}
	err := t.conn.WriteJSON(msg)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Next 实现TerminalSizeQueue接口
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.closeChan:
		return nil
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *PodExecHandler) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
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
	shell := c.DefaultQuery("shell", "/bin/sh")

	instanceID := uint(1)
	if instanceIdStr != "" {
		if id, err := strconv.ParseInt(instanceIdStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		sendExecError(conn, "K8s客户端未初始化")
		return
	}

	// 获取 RESTClient
	restClient := client.CoreV1().RESTClient()

	// 创建exec请求
	req := restClient.Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{shell},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 获取配置
	config, exists := configs.GetK8sConfig(instanceID)
	if !exists {
		sendExecError(conn, "获取K8s配置失败")
		return
	}

	// 创建executor
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		sendExecError(conn, "创建executor失败: "+err.Error())
		return
	}

	// 创建终端会话
	session := &TerminalSession{
		conn:      conn,
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		closeChan: make(chan struct{}),
		rows:      24,
		cols:      80,
	}
	defer close(session.closeChan)

	// 执行命令
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             session,
		Stdout:            session,
		Stderr:            session,
		Tty:               true,
		TerminalSizeQueue: session,
	})

	if err != nil {
		sendExecError(conn, "执行命令失败: "+err.Error())
	}
}

// sendExecError 发送错误消息
func sendExecError(conn *websocket.Conn, message string) {
	errorMsg := map[string]interface{}{
		"type":  "error",
		"error": message,
	}
	_ = conn.WriteJSON(errorMsg)
}
