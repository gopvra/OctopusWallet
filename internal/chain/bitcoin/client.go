package bitcoin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/octopuswallet/octopuswallet/internal/chain"
)

type Client struct {
	rpcURL   string
	rpcUser  string
	rpcPass  string
	network  *chaincfg.Params
	httpC    *http.Client
}

func NewClient(rpcURL, rpcUser, rpcPass, network string) (*Client, error) {
	var net *chaincfg.Params
	switch network {
	case "mainnet", "":
		net = &chaincfg.MainNetParams
	case "testnet":
		net = &chaincfg.TestNet3Params
	default:
		return nil, fmt.Errorf("unknown bitcoin network: %s", network)
	}
	return &Client{
		rpcURL:  rpcURL,
		rpcUser: rpcUser,
		rpcPass: rpcPass,
		network: net,
		httpC:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) Name() string           { return "bitcoin" }
func (c *Client) Type() chain.ChainType  { return chain.ChainTypeBTC }
func (c *Client) NativeSymbol() string    { return "BTC" }

func (c *Client) GetBalance(ctx context.Context, address string, token string) (string, error) {
	// Bitcoin doesn't have tokens; use scantxoutset or getreceivedbyaddress
	result, err := c.call(ctx, "getreceivedbyaddress", address, 1)
	if err != nil {
		// If address not in wallet, try scantxoutset
		return "0", nil
	}
	var btcAmount float64
	if err := json.Unmarshal(result, &btcAmount); err != nil {
		return "0", err
	}
	satoshi := int64(btcAmount * 1e8)
	return fmt.Sprintf("%d", satoshi), nil
}

func (c *Client) DeriveAddress(masterSeed []byte, merchantIndex, addressIndex uint32) (string, error) {
	return DeriveAddress(masterSeed, merchantIndex, addressIndex, c.network)
}

func (c *Client) DerivePrivateKey(masterSeed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	return DerivePrivateKey(masterSeed, merchantIndex, addressIndex)
}

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	if params == nil {
		params = []interface{}{}
	}
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.rpcURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.rpcUser != "" {
		req.SetBasicAuth(c.rpcUser, c.rpcPass)
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
	result, err := c.call(ctx, "getblockcount")
	if err != nil {
		return 0, err
	}
	var height uint64
	if err := json.Unmarshal(result, &height); err != nil {
		return 0, err
	}
	return height, nil
}

func (c *Client) ScanBlockForPayments(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]chain.IncomingTx, error) {
	// Get block hash
	hashResult, err := c.call(ctx, "getblockhash", blockHeight)
	if err != nil {
		return nil, err
	}
	var blockHash string
	if err := json.Unmarshal(hashResult, &blockHash); err != nil {
		return nil, err
	}

	// Get block with transactions (verbosity 2)
	blockResult, err := c.call(ctx, "getblock", blockHash, 2)
	if err != nil {
		return nil, err
	}

	var block struct {
		Tx []struct {
			TxID string `json:"txid"`
			Vin  []struct {
				TxID string `json:"txid"`
				Vout int    `json:"vout"`
			} `json:"vin"`
			Vout []struct {
				Value        float64 `json:"value"`
				ScriptPubKey struct {
					Address string `json:"address"`
				} `json:"scriptPubKey"`
			} `json:"vout"`
		} `json:"tx"`
	}
	if err := json.Unmarshal(blockResult, &block); err != nil {
		return nil, err
	}

	var txs []chain.IncomingTx
	for _, tx := range block.Tx {
		for _, vout := range tx.Vout {
			addr := vout.ScriptPubKey.Address
			if _, watched := watchAddresses[addr]; !watched {
				continue
			}
			// Convert BTC to satoshi
			satoshi := int64(vout.Value * 1e8)
			txs = append(txs, chain.IncomingTx{
				TxHash:      tx.TxID,
				FromAddress: "", // UTXO model - sender requires input resolution
				ToAddress:   addr,
				Amount:      fmt.Sprintf("%d", satoshi),
				Token:       "",
				BlockHeight: blockHeight,
			})
		}
	}

	return txs, nil
}

func (c *Client) GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error) {
	result, err := c.call(ctx, "getrawtransaction", txHash, true)
	if err != nil {
		return 0, err
	}
	var tx struct {
		Confirmations uint64 `json:"confirmations"`
	}
	if err := json.Unmarshal(result, &tx); err != nil {
		return 0, err
	}
	return tx.Confirmations, nil
}

func (c *Client) SendTransaction(ctx context.Context, req chain.SendRequest) (string, error) {
	// Bitcoin transaction construction requires UTXO selection, which is complex.
	// For MVP, we use Bitcoin Core's wallet RPC to handle UTXO management.
	result, err := c.call(ctx, "sendtoaddress", req.ToAddress, req.Amount)
	if err != nil {
		return "", fmt.Errorf("sendtoaddress: %w", err)
	}
	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return "", err
	}
	return txHash, nil
}

func (c *Client) EstimateFee(ctx context.Context, req chain.SendRequest) (string, error) {
	result, err := c.call(ctx, "estimatesmartfee", 6) // 6 blocks target
	if err != nil {
		return "", err
	}
	var estimate struct {
		FeeRate float64 `json:"feerate"`
	}
	if err := json.Unmarshal(result, &estimate); err != nil {
		return "", err
	}
	// Return fee rate in sat/vB (approximate for a standard tx ~250 vbytes)
	satPerVB := int64(estimate.FeeRate * 1e8 / 1000)
	fee := satPerVB * 250
	return fmt.Sprintf("%d", fee), nil
}
