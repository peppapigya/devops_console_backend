package websocket

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LogHub struct {
	clients map[uint64]map[*websocket.Conn]bool
	mu      sync.RWMutex
}

var hub = &LogHub{
	clients: make(map[uint64]map[*websocket.Conn]bool),
}

func HandleLogWebSocket(c *gin.Context) {
	executionIDStr := c.Param("id")
	executionID, err := strconv.ParseUint(executionIDStr, 10, 64)
	if err != nil {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hub.mu.Lock()
	if hub.clients[executionID] == nil {
		hub.clients[executionID] = make(map[*websocket.Conn]bool)
	}
	hub.clients[executionID][conn] = true
	hub.mu.Unlock()

	defer func() {
		hub.mu.Lock()
		delete(hub.clients[executionID], conn)
		if len(hub.clients[executionID]) == 0 {
			delete(hub.clients, executionID)
		}
		hub.mu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func BroadcastLog(executionID uint64, logStr string) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if clients, ok := hub.clients[executionID]; ok {
		for conn := range clients {
			err := conn.WriteMessage(websocket.TextMessage, []byte(logStr))
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
			}
		}
	}
}
