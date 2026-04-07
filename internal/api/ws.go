package api

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages WebSocket connections for real-time payment status updates.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{} // paymentID -> set of connections
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*websocket.Conn]struct{})}
}

// Subscribe adds a WebSocket connection for a payment.
func (h *Hub) Subscribe(paymentID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[paymentID] == nil {
		h.clients[paymentID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[paymentID][conn] = struct{}{}
}

// Unsubscribe removes a WebSocket connection.
func (h *Hub) Unsubscribe(paymentID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[paymentID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, paymentID)
		}
	}
}

// Broadcast sends a status update to all connections watching a payment.
func (h *Hub) Broadcast(paymentID string, message interface{}) {
	h.mu.RLock()
	conns := h.clients[paymentID]
	h.mu.RUnlock()

	for conn := range conns {
		if err := conn.WriteJSON(message); err != nil {
			slog.Debug("ws write error", "error", err)
			conn.Close()
			h.Unsubscribe(paymentID, conn)
		}
	}
}

// PaymentStatusUpdate is the message sent over WebSocket.
type PaymentStatusUpdate struct {
	PaymentID     string `json:"payment_id"`
	Status        string `json:"status"`
	TxHash        string `json:"tx_hash,omitempty"`
	Confirmations int    `json:"confirmations"`
	AmountReceived string `json:"amount_received,omitempty"`
}

// HandleWebSocket is the Gin handler for /ws/payments/:id
func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Param("id")
		if paymentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id required"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Error("ws upgrade failed", "error", err)
			return
		}
		defer conn.Close()

		hub.Subscribe(paymentID, conn)
		defer hub.Unsubscribe(paymentID, conn)

		// Keep connection alive, read pongs
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}
