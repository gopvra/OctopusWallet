package api

import (
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const maxConnsPerPayment = 10 // prevent connection flooding per payment

// Hub manages WebSocket connections for real-time payment status updates.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{} // paymentID -> set of connections
	total   atomic.Int64
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*websocket.Conn]struct{})}
}

// Subscribe adds a WebSocket connection for a payment.
func (h *Hub) Subscribe(paymentID string, conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[paymentID] == nil {
		h.clients[paymentID] = make(map[*websocket.Conn]struct{})
	}
	if len(h.clients[paymentID]) >= maxConnsPerPayment {
		return false
	}
	h.clients[paymentID][conn] = struct{}{}
	h.total.Add(1)
	return true
}

// Unsubscribe removes a WebSocket connection.
func (h *Hub) Unsubscribe(paymentID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[paymentID]; ok {
		if _, exists := conns[conn]; exists {
			delete(conns, conn)
			h.total.Add(-1)
		}
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
	PaymentID      string `json:"payment_id"`
	Status         string `json:"status"`
	TxHash         string `json:"tx_hash,omitempty"`
	Confirmations  int    `json:"confirmations"`
	AmountReceived string `json:"amount_received,omitempty"`
}

// newUpgrader creates a WebSocket upgrader that validates the Origin header
// against the provided list of allowed origins.
func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // non-browser clients (curl, etc.)
			}
			if originSet["*"] {
				return true
			}
			return originSet[origin]
		},
	}
}

// HandleWebSocket is the Gin handler for /ws/payments/:id
func HandleWebSocket(hub *Hub, allowedOrigins []string) gin.HandlerFunc {
	upgrader := newUpgrader(allowedOrigins)

	return func(c *gin.Context) {
		paymentID := c.Param("id")

		// Validate payment ID is a proper UUID to prevent abuse
		if _, err := uuid.Parse(paymentID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			slog.Error("ws upgrade failed", "error", err)
			return
		}
		defer conn.Close()

		if !hub.Subscribe(paymentID, conn) {
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "too many connections"))
			return
		}
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
