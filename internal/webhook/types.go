package webhook

import "time"

type EventType string

const (
	EventPaymentConfirming EventType = "payment.confirming"
	EventPaymentCompleted  EventType = "payment.completed"
	EventPaymentExpired    EventType = "payment.expired"
	EventPayoutCompleted   EventType = "payout.completed"
	EventPayoutFailed      EventType = "payout.failed"
)

type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

type PaymentEventData struct {
	PaymentID      string `json:"payment_id"`
	Chain          string `json:"chain"`
	Address        string `json:"address"`
	AmountExpected string `json:"amount_expected"`
	AmountReceived string `json:"amount_received"`
	TxHash         string `json:"tx_hash,omitempty"`
	Confirmations  int    `json:"confirmations"`
	Status         string `json:"status"`
}

type PayoutEventData struct {
	PayoutID  string `json:"payout_id"`
	Chain     string `json:"chain"`
	ToAddress string `json:"to_address"`
	Amount    string `json:"amount"`
	TxHash    string `json:"tx_hash,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}
