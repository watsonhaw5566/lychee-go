package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"lychee-go/internal/config"
	"lychee-go/internal/logger"
)

var upgrader *websocket.Upgrader

func Init() {
	readBufferSize := config.GetInt("websocket.read_buffer_size", 1024)
	writeBufferSize := config.GetInt("websocket.write_buffer_size", 1024)

	upgrader = &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			if config.IsSet("websocket.allow_origins") {
				origins := config.GetString("websocket.allow_origins", "")
				return origins == "*" || origins == ""
			}
			return true
		},
	}

	go hub()

	logger.Info("[WebSocket] Initialized with read_buffer=%d, write_buffer=%d", readBufferSize, writeBufferSize)
}

var (
	clients    = make(map[*websocket.Conn]bool)
	broadcast  = make(chan Message)
	register   = make(chan *websocket.Conn)
	unregister = make(chan *websocket.Conn)
	clientMu   sync.Mutex
	handlers   = make(map[string]MessageHandler)
	handlersMu sync.RWMutex
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MessageHandler func(conn *websocket.Conn, message Message) error

func hub() {
	for {
		select {
		case conn := <-register:
			clientMu.Lock()
			clients[conn] = true
			clientMu.Unlock()
			logger.Info("[WebSocket] Client connected, total: %d", len(clients))

		case conn := <-unregister:
			clientMu.Lock()
			if _, ok := clients[conn]; ok {
				delete(clients, conn)
				conn.Close()
			}
			clientMu.Unlock()
			logger.Info("[WebSocket] Client disconnected, total: %d", len(clients))

		case message := <-broadcast:
			clientMu.Lock()
			for conn := range clients {
				err := conn.WriteJSON(message)
				if err != nil {
					logger.Error("[WebSocket] Write error: %v", err)
					conn.Close()
					delete(clients, conn)
				}
			}
			clientMu.Unlock()
		}
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("[WebSocket] Upgrade error: %v", err)
		return
	}

	register <- conn

	defer func() {
		unregister <- conn
	}()

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Info("[WebSocket] Client closed connection normally")
			} else {
				logger.Error("[WebSocket] Read error: %v", err)
			}
			break
		}

		handlersMu.RLock()
		handler, ok := handlers[message.Type]
		handlersMu.RUnlock()

		if ok {
			if err := handler(conn, message); err != nil {
				logger.Error("[WebSocket] Handler error for type %s: %v", message.Type, err)
				sendError(conn, err.Error())
			}
		} else {
			logger.Warn("[WebSocket] No handler registered for message type: %s", message.Type)
			sendError(conn, fmt.Sprintf("Unknown message type: %s", message.Type))
		}
	}
}

func RegisterHandler(messageType string, handler MessageHandler) {
	handlersMu.Lock()
	handlers[messageType] = handler
	handlersMu.Unlock()
	logger.Info("[WebSocket] Registered handler for message type: %s", messageType)
}

func Broadcast(messageType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	broadcast <- Message{
		Type:    messageType,
		Payload: data,
	}

	return nil
}

func Send(conn *websocket.Conn, messageType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	return conn.WriteJSON(Message{
		Type:    messageType,
		Payload: data,
	})
}

func sendError(conn *websocket.Conn, message string) {
	err := conn.WriteJSON(Message{
		Type:    "error",
		Payload: json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, message)),
	})
	if err != nil {
		logger.Error("[WebSocket] Send error failed: %v", err)
	}
}

func GetClientCount() int {
	clientMu.Lock()
	defer clientMu.Unlock()
	return len(clients)
}
