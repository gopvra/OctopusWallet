package chain

import "context"

type ChainType string

const (
	ChainTypeEVM    ChainType = "evm"
	ChainTypeSolana ChainType = "solana"
	ChainTypeTron   ChainType = "tron"
	ChainTypeBTC    ChainType = "bitcoin"
)

type Chain interface {
	// Identity
	Name() string
	Type() ChainType
	NativeSymbol() string

	// Wallet
	DeriveAddress(masterSeed []byte, merchantIndex, addressIndex uint32) (string, error)
	DerivePrivateKey(masterSeed []byte, merchantIndex, addressIndex uint32) ([]byte, error)

	// Monitoring
	GetCurrentBlockHeight(ctx context.Context) (uint64, error)
	ScanBlockForPayments(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]IncomingTx, error)
	GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error)

	// Balance
	GetBalance(ctx context.Context, address string, token string) (string, error)

	// Sending
	SendTransaction(ctx context.Context, req SendRequest) (string, error)
	EstimateFee(ctx context.Context, req SendRequest) (string, error)
}

type IncomingTx struct {
	TxHash      string
	FromAddress string
	ToAddress   string
	Amount      string // decimal string in base units
	Token       string // empty for native, contract address for tokens
	BlockHeight uint64
}

type SendRequest struct {
	FromAddress string
	ToAddress   string
	Amount      string
	Token       string // empty for native
	PrivateKey  []byte
}
