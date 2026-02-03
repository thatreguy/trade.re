package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

// MessageType identifies the kind of WebSocket message
type MessageType string

const (
	TypeTrade        MessageType = "trade"
	TypeOrderBook    MessageType = "orderbook"
	TypePosition     MessageType = "position"
	TypeOrder        MessageType = "order"
	TypeOI           MessageType = "oi"
	TypeLiquidation  MessageType = "liquidation"
	TypeSubscribe    MessageType = "subscribe"
	TypeUnsubscribe  MessageType = "unsubscribe"
)

// Message is the WebSocket message envelope
type Message struct {
	Type       MessageType `json:"type"`
	Channel    string      `json:"channel,omitempty"`
	Data       interface{} `json:"data"`
	Timestamp  int64       `json:"timestamp"`
}

// Client represents a WebSocket connection
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool
	mu            sync.RWMutex
}

// Hub manages all WebSocket clients and broadcasts
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToChannel sends a message to clients subscribed to a channel
func (h *Hub) BroadcastToChannel(channel string, msg Message) {
	msg.Timestamp = time.Now().UnixMilli()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		client.mu.RLock()
		subscribed := client.subscriptions[channel]
		client.mu.RUnlock()

		if subscribed {
			select {
			case client.send <- data:
			default:
				// Client buffer full, skip
			}
		}
	}
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(msg Message) {
	msg.Timestamp = time.Now().UnixMilli()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	h.broadcast <- data
}

// BroadcastTrade sends a trade to all clients (trades are always public)
func (h *Hub) BroadcastTrade(trade interface{}) {
	h.Broadcast(Message{
		Type: TypeTrade,
		Data: trade,
	})
}

// BroadcastOrderBook sends order book update
func (h *Hub) BroadcastOrderBook(instrument string, book interface{}) {
	h.BroadcastToChannel("orderbook:"+instrument, Message{
		Type:    TypeOrderBook,
		Channel: "orderbook:" + instrument,
		Data:    book,
	})
}

// BroadcastPosition sends position update (positions are public)
func (h *Hub) BroadcastPosition(position interface{}) {
	h.Broadcast(Message{
		Type: TypePosition,
		Data: position,
	})
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
}

// Subscribe adds a channel subscription
func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	c.subscriptions[channel] = true
	c.mu.Unlock()
}

// Unsubscribe removes a channel subscription
func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	delete(c.subscriptions, channel)
	c.mu.Unlock()
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle subscription messages
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case TypeSubscribe:
			if channel, ok := msg.Data.(string); ok {
				c.Subscribe(channel)
			}
		case TypeUnsubscribe:
			if channel, ok := msg.Data.(string); ok {
				c.Unsubscribe(channel)
			}
		}
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch pending messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
