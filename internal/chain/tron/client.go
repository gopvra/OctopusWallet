package tron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/octopuswallet/octopuswallet/internal/chain"
)

type Client struct {
	rpcURL string
	apiKey string
	httpC  *http.Client
}

func NewClient(rpcURL, apiKey string) (*Client, error) {
	return &Client{
		rpcURL: rpcURL,
		apiKey: apiKey,
		httpC:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) Name() string           { return "tron" }
func (c *Client) Type() chain.ChainType  { return chain.ChainTypeTron }
func (c *Client) NativeSymbol() string    { return "TRX" }

func (c *Client) GetBalance(ctx context.Context, address string, token string) (string, error) {
	if token == "" {
		data, err := c.apiCall(ctx, "/wallet/getaccount", map[string]string{"address": address})
		if err != nil {
			return "0", err
		}
		var account struct {
			Balance int64 `json:"balance"`
		}
		if err := json.Unmarshal(data, &account); err != nil {
			return "0", err
		}
		return fmt.Sprintf("%d", account.Balance), nil
	}
	// TRC-20 balance via triggerconstantcontract
	// balanceOf(address) selector = 70a08231
	addrHex, err := TronAddressToHex(address)
	if err != nil {
		return "0", err
	}
	// Pad to 32 bytes
	paddedAddr := fmt.Sprintf("%064s", addrHex[2:]) // remove 41 prefix, pad to 64 hex
	data, err := c.apiCall(ctx, "/wallet/triggerconstantcontract", map[string]interface{}{
		"owner_address":     address,
		"contract_address":  token,
		"function_selector": "balanceOf(address)",
		"parameter":         paddedAddr,
	})
	if err != nil {
		return "0", err
	}
	var resp struct {
		ConstantResult []string `json:"constant_result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "0", err
	}
	if len(resp.ConstantResult) > 0 {
		balance := new(big.Int)
		balance.SetString(resp.ConstantResult[0], 16)
		return balance.String(), nil
	}
	return "0", nil
}

func (c *Client) DeriveAddress(masterSeed []byte, merchantIndex, addressIndex uint32) (string, error) {
	return DeriveAddress(masterSeed, merchantIndex, addressIndex)
}

func (c *Client) DerivePrivateKey(masterSeed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	return DerivePrivateKey(masterSeed, merchantIndex, addressIndex)
}

func (c *Client) apiCall(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.rpcURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpC.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) GetCurrentBlockHeight(ctx context.Context) (uint64, error) {
	data, err := c.apiCall(ctx, "/wallet/getnowblock", nil)
	if err != nil {
		return 0, err
	}

	var block struct {
		BlockHeader struct {
			RawData struct {
				Number uint64 `json:"number"`
			} `json:"raw_data"`
		} `json:"block_header"`
	}
	if err := json.Unmarshal(data, &block); err != nil {
		return 0, err
	}
	return block.BlockHeader.RawData.Number, nil
}

func (c *Client) ScanBlockForPayments(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]chain.IncomingTx, error) {
	data, err := c.apiCall(ctx, "/wallet/getblockbynum", map[string]uint64{"num": blockHeight})
	if err != nil {
		return nil, err
	}

	var block struct {
		Transactions []struct {
			TxID    string `json:"txID"`
			RawData struct {
				Contract []struct {
					Type      string `json:"type"`
					Parameter struct {
						Value struct {
							Amount       int64  `json:"amount"`
							OwnerAddress string `json:"owner_address"`
							ToAddress    string `json:"to_address"`
							// TRC-20 fields
							ContractAddress string `json:"contract_address"`
							Data            string `json:"data"`
						} `json:"value"`
					} `json:"parameter"`
				} `json:"contract"`
			} `json:"raw_data"`
		} `json:"transactions"`
	}
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("parse tron block: %w", err)
	}

	var txs []chain.IncomingTx
	for _, tx := range block.Transactions {
		for _, contract := range tx.RawData.Contract {
			if contract.Type == "TransferContract" {
				toAddr := hexToTronAddress(contract.Parameter.Value.ToAddress)
				if _, watched := watchAddresses[toAddr]; !watched {
					continue
				}
				fromAddr := hexToTronAddress(contract.Parameter.Value.OwnerAddress)
				txs = append(txs, chain.IncomingTx{
					TxHash:      tx.TxID,
					FromAddress: fromAddr,
					ToAddress:   toAddr,
					Amount:      fmt.Sprintf("%d", contract.Parameter.Value.Amount),
					Token:       "",
					BlockHeight: blockHeight,
				})
			}
			if contract.Type == "TriggerSmartContract" {
				// TRC-20 transfer detection
				trc20Tx := parseTRC20Transfer(tx.TxID, contract.Parameter.Value.ContractAddress,
					contract.Parameter.Value.Data, contract.Parameter.Value.OwnerAddress, blockHeight, watchAddresses)
				if trc20Tx != nil {
					txs = append(txs, *trc20Tx)
				}
			}
		}
	}

	return txs, nil
}

func (c *Client) GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error) {
	data, err := c.apiCall(ctx, "/wallet/gettransactioninfobyid", map[string]string{"value": txHash})
	if err != nil {
		return 0, err
	}

	var txInfo struct {
		BlockNumber uint64 `json:"blockNumber"`
	}
	if err := json.Unmarshal(data, &txInfo); err != nil {
		return 0, err
	}

	currentHeight, err := c.GetCurrentBlockHeight(ctx)
	if err != nil {
		return 0, err
	}

	if currentHeight <= txInfo.BlockNumber {
		return 0, nil
	}
	return currentHeight - txInfo.BlockNumber + 1, nil
}

func (c *Client) SendTransaction(ctx context.Context, req chain.SendRequest) (string, error) {
	// Create transaction
	createData, err := c.apiCall(ctx, "/wallet/createtransaction", map[string]interface{}{
		"to_address":    req.ToAddress,
		"owner_address": req.FromAddress,
		"amount":        req.Amount,
	})
	if err != nil {
		return "", fmt.Errorf("create tron transaction: %w", err)
	}

	// Sign and broadcast would require signing with the private key
	// For now, return the raw transaction for external signing
	var tx struct {
		TxID string `json:"txID"`
	}
	if err := json.Unmarshal(createData, &tx); err != nil {
		return "", err
	}

	return tx.TxID, fmt.Errorf("tron transaction signing not yet implemented")
}

func (c *Client) EstimateFee(ctx context.Context, req chain.SendRequest) (string, error) {
	// TRON native TRX transfers consume bandwidth, which is free up to a daily limit.
	// TRC-20 transfers consume energy. Base fee for simple TRX transfer is ~0.
	return "0", nil
}
