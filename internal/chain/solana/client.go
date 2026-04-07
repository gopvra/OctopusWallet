package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/chain"
)

type Client struct {
	rpcURL string
	httpC  *http.Client
}

func NewClient(rpcURL string) (*Client, error) {
	return &Client{
		rpcURL: rpcURL,
		httpC:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) Name() string           { return "solana" }
func (c *Client) Type() chain.ChainType  { return chain.ChainTypeSolana }
func (c *Client) NativeSymbol() string    { return "SOL" }

func (c *Client) GetBalance(ctx context.Context, address string, token string) (string, error) {
	if token == "" {
		result, err := c.call(ctx, "getBalance", address)
		if err != nil {
			return "0", err
		}
		var resp struct {
			Value uint64 `json:"value"`
		}
		if err := json.Unmarshal(result, &resp); err != nil {
			return "0", err
		}
		return fmt.Sprintf("%d", resp.Value), nil
	}
	// SPL token balance
	result, err := c.call(ctx, "getTokenAccountsByOwner", address,
		map[string]string{"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
		map[string]string{"encoding": "jsonParsed"})
	if err != nil {
		return "0", err
	}
	var resp struct {
		Value []struct {
			Account struct {
				Data struct {
					Parsed struct {
						Info struct {
							TokenAmount struct {
								Amount string `json:"amount"`
							} `json:"tokenAmount"`
						} `json:"info"`
					} `json:"parsed"`
				} `json:"data"`
			} `json:"account"`
		} `json:"value"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return "0", err
	}
	if len(resp.Value) > 0 {
		return resp.Value[0].Account.Data.Parsed.Info.TokenAmount.Amount, nil
	}
	return "0", nil
}

func (c *Client) DeriveAddress(masterSeed []byte, merchantIndex, addressIndex uint32) (string, error) {
	return DeriveAddress(masterSeed, merchantIndex, addressIndex)
}

func (c *Client) DerivePrivateKey(masterSeed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	return DerivePrivateKey(masterSeed, merchantIndex, addressIndex)
}

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.rpcURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpC.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, fmt.Errorf("parse rpc response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func (c *Client) GetCurrentBlockHeight(ctx context.Context) (uint64, error) {
	result, err := c.call(ctx, "getSlot")
	if err != nil {
		return 0, err
	}
	var slot uint64
	if err := json.Unmarshal(result, &slot); err != nil {
		return 0, err
	}
	return slot, nil
}

func (c *Client) ScanBlockForPayments(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]chain.IncomingTx, error) {
	result, err := c.call(ctx, "getBlock", blockHeight, map[string]interface{}{
		"encoding":                       "jsonParsed",
		"transactionDetails":             "full",
		"maxSupportedTransactionVersion": 0,
	})
	if err != nil {
		return nil, err
	}

	var block struct {
		Transactions []struct {
			Transaction struct {
				Signatures []string `json:"signatures"`
				Message    struct {
					AccountKeys []struct {
						Pubkey string `json:"pubkey"`
					} `json:"accountKeys"`
					Instructions []struct {
						Program  string `json:"program"`
						Parsed   *struct {
							Type string `json:"type"`
							Info struct {
								Destination string `json:"destination"`
								Source      string `json:"source"`
								Lamports   uint64 `json:"lamports"`
								Amount     string `json:"amount"`
							} `json:"info"`
						} `json:"parsed,omitempty"`
					} `json:"instructions"`
				} `json:"message"`
			} `json:"transaction"`
			Meta *struct {
				Err interface{} `json:"err"`
			} `json:"meta"`
		} `json:"transactions"`
	}

	if err := json.Unmarshal(result, &block); err != nil {
		return nil, fmt.Errorf("parse block: %w", err)
	}

	var txs []chain.IncomingTx
	for _, txn := range block.Transactions {
		if txn.Meta != nil && txn.Meta.Err != nil {
			continue // skip failed transactions
		}
		for _, inst := range txn.Transaction.Message.Instructions {
			if inst.Parsed == nil {
				continue
			}
			if inst.Program == "system" && inst.Parsed.Type == "transfer" {
				dest := inst.Parsed.Info.Destination
				if _, watched := watchAddresses[dest]; watched {
					txs = append(txs, chain.IncomingTx{
						TxHash:      txn.Transaction.Signatures[0],
						FromAddress: inst.Parsed.Info.Source,
						ToAddress:   dest,
						Amount:      fmt.Sprintf("%d", inst.Parsed.Info.Lamports),
						Token:       "",
						BlockHeight: blockHeight,
					})
				}
			}
			// SPL Token transfers
			if inst.Program == "spl-token" && inst.Parsed.Type == "transfer" {
				dest := inst.Parsed.Info.Destination
				if _, watched := watchAddresses[dest]; watched {
					txs = append(txs, chain.IncomingTx{
						TxHash:      txn.Transaction.Signatures[0],
						FromAddress: inst.Parsed.Info.Source,
						ToAddress:   dest,
						Amount:      inst.Parsed.Info.Amount,
						Token:       "spl-token",
						BlockHeight: blockHeight,
					})
				}
			}
		}
	}

	return txs, nil
}

func (c *Client) GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error) {
	result, err := c.call(ctx, "getTransaction", txHash, map[string]string{"encoding": "json"})
	if err != nil {
		return 0, err
	}

	var tx struct {
		Slot uint64 `json:"slot"`
	}
	if err := json.Unmarshal(result, &tx); err != nil {
		return 0, err
	}

	currentSlot, err := c.GetCurrentBlockHeight(ctx)
	if err != nil {
		return 0, err
	}

	if currentSlot <= tx.Slot {
		return 0, nil
	}
	return currentSlot - tx.Slot, nil
}

func (c *Client) SendTransaction(ctx context.Context, req chain.SendRequest) (string, error) {
	// Solana transaction construction requires recent blockhash and serialization.
	// This is a simplified version - production would use a proper Solana SDK.
	return "", fmt.Errorf("solana SendTransaction: not yet implemented - requires transaction serialization")
}

func (c *Client) EstimateFee(ctx context.Context, req chain.SendRequest) (string, error) {
	// Solana has a fixed base fee of 5000 lamports per signature
	return "5000", nil
}
