package webhook

import "time"

type EventType string

const (
	EventPaymentConfirming EventType = "payment.confirming"
	EventPaymentCompleted  EventType = "payment.completed"
	EventPaymentExpired    EventType = "payment.expired"
	EventPayoutCompleted        EventType = "payout.completed"
	EventPayoutFailed           EventType = "payout.failed"
	EventPayoutPendingApproval  EventType = "payout.pending_approval"
	EventPayoutApproved         EventType = "payout.approved"
	EventPayoutRejected         EventType = "payout.rejected"
	EventSweepCompleted         EventType = "sweep.completed"
	EventSweepFailed            EventType = "sweep.failed"
	EventTransferCompleted      EventType = "transfer.completed"
	EventTransferFailed         EventType = "transfer.failed"
	EventRefundCompleted        EventType = "refund.completed"
	EventRefundFailed           EventType = "refund.failed"
	EventGasDepositCompleted    EventType = "gas.deposit_completed"
	EventGasLowBalance          EventType = "gas.low_balance"
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

type SweepEventData struct {
	SweepTaskID string `json:"sweep_task_id"`
	PaymentID   string `json:"payment_id,omitempty"`
	Chain       string `json:"chain"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	TxHash      string `json:"tx_hash,omitempty"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

type TransferEventData struct {
	TransferID   string `json:"transfer_id"`
	Chain        string `json:"chain"`
	TransferType string `json:"transfer_type"`
	FromAddress  string `json:"from_address"`
	ToAddress    string `json:"to_address"`
	Amount       string `json:"amount"`
	TxHash       string `json:"tx_hash,omitempty"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
}

type ApprovalEventData struct {
	PayoutID string `json:"payout_id"`
	Chain    string `json:"chain"`
	Amount   string `json:"amount"`
	Status   string `json:"status"`
	Approver string `json:"approver,omitempty"`
}

type GasEventData struct {
	Chain          string `json:"chain"`
	StationAddress string `json:"station_address,omitempty"`
	ToAddress      string `json:"to_address,omitempty"`
	Amount         string `json:"amount"`
	Balance        string `json:"balance,omitempty"`
	TxHash         string `json:"tx_hash,omitempty"`
	Status         string `json:"status"`
}
