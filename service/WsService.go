package service

import (
	"fmt"
	"lychee/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

// WsManager WebSocket 管理器
var WsManager = clientManager{
	clientGroup: make(map[string]map[string]*wsClient),
	register:    make(chan *wsClient),
	unRegister:  make(chan *wsClient),
	broadcast:   make(chan *boradcastData, 10),
}

// ClientManager websocket client Manager struct
type clientManager struct {
	clientGroup map[string]map[string]*wsClient
	register    chan *wsClient
	unRegister  chan *wsClient
	broadcast   chan *boradcastData
}

// boradcastData 广播数据
type boradcastData struct {
	GroupID string
	Data    []byte
}

// wsClient Websocket 客户端
type wsClient struct {
	ID     string
	Group  string
	Socket *websocket.Conn
	Send   chan []byte
}

func (c *wsClient) Read() {
	defer func() {
		WsManager.unRegister <- c
		c.Socket.Close()
	}()

	for {
		_, _, err := c.Socket.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *wsClient) Write() {
	defer func() {
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Socket.WriteMessage(websocket.BinaryMessage, message)
		}
	}
}

// Start 启动 websocket 管理器
func (manager *clientManager) Start() {
	logger.ZapLogger.Info("Websocket manage start")
	for {
		select {
		case client := <-manager.register:
			logger.ZapLogger.Info(fmt.Sprintf("Websocket client %s connect", client.ID))
			if manager.clientGroup[client.Group] == nil {
				manager.clientGroup[client.Group] = make(map[string]*wsClient)
			}
			manager.clientGroup[client.Group][client.ID] = client
			logger.ZapLogger.Info(fmt.Sprintf("Register client %s to %s group success", client.ID, client.Group))

		case client := <-manager.unRegister:
			logger.ZapLogger.Info(fmt.Sprintf("Unregister websocket client %s", client.ID))
			if _, ok := manager.clientGroup[client.Group]; ok {
				if _, ok := manager.clientGroup[client.Group][client.ID]; ok {
					close(client.Send)
					delete(manager.clientGroup[client.Group], client.ID)
					logger.ZapLogger.Info(fmt.Sprintf("Unregister websocket client %s from group %s success", client.ID, client.Group))

					if len(manager.clientGroup[client.Group]) == 0 {
						logger.ZapLogger.Info(fmt.Sprintf("Clear no client group %s", client.Group))
						delete(manager.clientGroup, client.Group)
					}
				}
			}

		case data := <-manager.broadcast:
			if groupMap, ok := manager.clientGroup[data.GroupID]; ok {
				for _, conn := range groupMap {
					conn.Send <- data.Data
				}
			}
		}
	}
}

// RegisterClient 向 manage 中注册 client
func (manager *clientManager) RegisterClient(ctx *gin.Context) {
	upgrader := websocket.Upgrader{
		// cross origin domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		// 处理 Sec-WebSocket-Protocol Header
		Subprotocols: []string{ctx.GetHeader("Sec-WebSocket-Protocol")},
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.ZapLogger.Error(fmt.Sprintf("websocket client connect %v error", ctx.Param("channel")))
		return
	}

	client := &wsClient{
		ID:     uuid.NewV4().String(),
		Group:  ctx.Param("channel"),
		Socket: conn,
		Send:   make(chan []byte, 1024),
	}

	manager.register <- client
	go client.Read()
	go client.Write()
}

// Groupbroadcast 向指定的 Group 广播
func (manager *clientManager) Groupbroadcast(group string, message []byte) {
	data := &boradcastData{
		GroupID: group,
		Data:    message,
	}
	manager.broadcast <- data
}
