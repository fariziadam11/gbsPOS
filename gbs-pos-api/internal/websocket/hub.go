package websocket

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	ClientPOS       = "pos"
	ClientCompanion = "companion"
)

type Client struct {
	Type      string
	ID        string
	conn      *websocket.Conn
	send      chan []byte
	hub       *Hub
	closeOnce sync.Once
}

type Hub struct {
	mu        sync.RWMutex
	clients   map[string]map[*Client]struct{}
	onMessage func(*Client, Message)
	upgrader  websocket.Upgrader
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
	}
}

func (h *Hub) SetAllowedOrigins(value string) {
	allowed := make(map[string]struct{})
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			allowed[origin] = struct{}{}
		}
	}
	h.upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" || len(allowed) == 0 {
			return true
		}
		_, ok := allowed[origin]
		return ok
	}
}

func (h *Hub) SetMessageHandler(handler func(*Client, Message)) {
	h.onMessage = handler
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, clientType, id string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{Type: clientType, ID: id, conn: conn, send: make(chan []byte, 16), hub: h}
	h.register(client)

	go client.writePump()
	client.readPump()
}

func (h *Hub) Send(clientType, id string, message Message) {
	data, err := jsonMarshal(message)
	if err != nil {
		return
	}
	key := clientType + ":" + id
	h.mu.RLock()
	clients := h.clients[key]
	for client := range clients {
		select {
		case client.send <- data:
		default:
			go client.close()
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := client.Type + ":" + client.ID
	if h.clients[key] == nil {
		h.clients[key] = make(map[*Client]struct{})
	}
	h.clients[key][client] = struct{}{}
	client.sendMessage(Message{Type: "CONNECTION_ACK", Message: "connected"})
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := client.Type + ":" + client.ID
	if clients := h.clients[key]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, key)
		}
	}
}

func (c *Client) sendMessage(message Message) {
	data, err := jsonMarshal(message)
	if err == nil {
		c.send <- data
	}
}

func (c *Client) readPump() {
	defer c.close()
	c.conn.SetReadLimit(64 * 1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		message, err := DecodeMessage(data)
		if err == nil && c.hub.onMessage != nil {
			c.hub.onMessage(c, message)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer c.close()
	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	})
}

func jsonMarshal(message Message) ([]byte, error) {
	return json.Marshal(message)
}
