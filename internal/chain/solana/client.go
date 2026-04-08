package solana

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	sol "github.com/cielu/go-solana"
	"github.com/cielu/go-solana/core/spltoken"
	solsys "github.com/cielu/go-solana/core/system"
	soltoken "github.com/cielu/go-solana/core/token"
	"github.com/mr-tron/base58"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
)

const coinTypeSOL = 501

type Client struct {
	sc     *sol.Client
	rpcURL string
}

func NewClient(rpcURL string) (*Client, error) {
	sc, err := sol.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("solana dial: %w", err)
	}
	return &Client{sc: sc, rpcURL: rpcURL}, nil
}

func (c *Client) Name() string          { return "solana" }
func (c *Client) Type() chain.ChainType { return chain.ChainTypeSolana }
func (c *Client) NativeSymbol() string   { return "SOL" }

// --- Balance ---

func (c *Client) GetBalance(ctx context.Context, address string, token string) (string, error) {
	pubkey := sol.Base58ToPublicKey(address)

	if token == "" {
		bal, err := c.sc.GetBalance(ctx, pubkey)
		if err != nil {
			return "0", err
		}
		return bal.Balance.String(), nil
	}

	// SPL token balance
	mint := sol.Base58ToPublicKey(token)
	accounts, err := c.sc.GetTokenAccountsByOwner(ctx, pubkey, sol.RpcMintWithProgramID{Mint: &mint})
	if err != nil {
		return "0", err
	}
	if len(accounts.Accounts) > 0 {
		tokenBal, err := c.sc.GetTokenAccountBalance(ctx, accounts.Accounts[0].Pubkey)
		if err != nil {
			return "0", err
		}
		return tokenBal.UiToken.Amount, nil
	}
	return "0", nil
}

// --- Address Derivation ---

func (c *Client) DeriveAddress(masterSeed []byte, merchantIndex, addressIndex uint32) (string, error) {
	key, err := deriveSolanaKey(masterSeed, merchantIndex, addressIndex)
	if err != nil {
		return "", err
	}
	pubKey := ed25519.PrivateKey(key).Public().(ed25519.PublicKey)
	return base58.Encode(pubKey), nil
}

func (c *Client) DerivePrivateKey(masterSeed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	return deriveSolanaKey(masterSeed, merchantIndex, addressIndex)
}

func deriveSolanaKey(seed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	key, err := wallet.DeriveKey(seed, 44, coinTypeSOL, merchantIndex, 0, addressIndex)
	if err != nil {
		return nil, fmt.Errorf("derive solana key: %w", err)
	}
	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("get private key: %w", err)
	}
	rawKey := privKey.Serialize()
	ed25519Seed := make([]byte, ed25519.SeedSize)
	copy(ed25519Seed, rawKey)
	return ed25519.NewKeyFromSeed(ed25519Seed), nil
}

// --- Block Scanning ---

func (c *Client) GetCurrentBlockHeight(ctx context.Context) (uint64, error) {
	return c.sc.GetSlot(ctx)
}

func (c *Client) ScanBlockForPayments(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]chain.IncomingTx, error) {
	// Use raw JSON-RPC for jsonParsed encoding (SDK typed structs don't support parsed instructions)
	result, err := c.rpcCall(ctx, "getBlock", blockHeight, map[string]interface{}{
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
					Instructions []struct {
						Program string `json:"program"`
						Parsed  *struct {
							Type string `json:"type"`
							Info struct {
								Destination string  `json:"destination"`
								Source      string  `json:"source"`
								Lamports    float64 `json:"lamports"`
								Amount      string  `json:"amount"`
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
			continue
		}
		for _, inst := range txn.Transaction.Message.Instructions {
			if inst.Parsed == nil {
				continue
			}
			dest := inst.Parsed.Info.Destination
			source := inst.Parsed.Info.Source
			if _, watched := watchAddresses[dest]; !watched {
				continue
			}
			sig := ""
			if len(txn.Transaction.Signatures) > 0 {
				sig = txn.Transaction.Signatures[0]
			}
			if inst.Program == "system" && inst.Parsed.Type == "transfer" {
				txs = append(txs, chain.IncomingTx{
					TxHash: sig, FromAddress: source, ToAddress: dest,
					Amount: fmt.Sprintf("%.0f", inst.Parsed.Info.Lamports),
					Token: "", BlockHeight: blockHeight,
				})
			}
			if inst.Program == "spl-token" && inst.Parsed.Type == "transfer" {
				txs = append(txs, chain.IncomingTx{
					TxHash: sig, FromAddress: source, ToAddress: dest,
					Amount: inst.Parsed.Info.Amount, Token: "spl-token", BlockHeight: blockHeight,
				})
			}
		}
	}
	return txs, nil
}

// rpcCall is a low-level JSON-RPC helper for endpoints that need raw JSON responses
// (e.g. jsonParsed block encoding not supported by typed SDK structs).
func (c *Client) rpcCall(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	body, _ := json.Marshal(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	req, err := http.NewRequestWithContext(ctx, "POST", c.rpcURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// --- Transaction Confirmation ---

func (c *Client) GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error) {
	sig := sol.Base58ToSignature(txHash)
	txInfo, err := c.sc.GetTransaction(ctx, sig)
	if err != nil {
		return 0, err
	}

	currentSlot, err := c.sc.GetSlot(ctx)
	if err != nil {
		return 0, err
	}

	if currentSlot <= txInfo.Slot {
		return 0, nil
	}
	return currentSlot - txInfo.Slot, nil
}

// --- Send Transaction (SOL native + SPL Token) ---

func (c *Client) SendTransaction(ctx context.Context, req chain.SendRequest) (string, error) {
	amount := new(big.Int)
	if _, ok := amount.SetString(req.Amount, 10); !ok {
		return "", fmt.Errorf("invalid amount: %s", req.Amount)
	}

	fromPubkey := sol.Base58ToPublicKey(req.FromAddress)
	toPubkey := sol.Base58ToPublicKey(req.ToAddress)

	var instructions []sol.Instruction

	if req.Token == "" {
		// --- SOL native transfer ---
		inst := solsys.NewTransferInstruction(fromPubkey, toPubkey, amount.Uint64())
		instructions = append(instructions, inst.Build())
	} else {
		// --- SPL Token transfer ---
		mintPubkey := sol.Base58ToPublicKey(req.Token)

		// Derive Associated Token Accounts for sender and recipient
		srcATA, _, err := spltoken.FindAssociatedTokenAddress(fromPubkey, mintPubkey)
		if err != nil {
			return "", fmt.Errorf("find source ATA: %w", err)
		}
		destATA, _, err := spltoken.FindAssociatedTokenAddress(toPubkey, mintPubkey)
		if err != nil {
			return "", fmt.Errorf("find destination ATA: %w", err)
		}

		// Build TransferChecked instruction
		transferInst := soltoken.NewTransferCheckedInstructionBuilder().
			SetAmount(amount.Uint64()).
			SetDecimals(req.Decimals).
			SetSourceAccount(srcATA).
			SetMintAccount(mintPubkey).
			SetDestinationAccount(destATA).
			SetOwnerAccount(fromPubkey)

		instructions = append(instructions, transferInst.Build())
	}

	// Get recent blockhash
	blockHashResult, err := c.sc.GetLatestBlockhash(ctx)
	if err != nil {
		return "", fmt.Errorf("get recent blockhash: %w", err)
	}

	// Create transaction
	tx, err := sol.NewTransaction(instructions, blockHashResult.LastBlock.Blockhash, fromPubkey)
	if err != nil {
		return "", fmt.Errorf("create transaction: %w", err)
	}

	// Sign with ed25519 private key
	privKey := ed25519.PrivateKey(req.PrivateKey)
	msgBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("marshal message: %w", err)
	}
	sigBytes := ed25519.Sign(privKey, msgBytes)
	var txSig sol.Signature
	copy(txSig[:], sigBytes)
	tx.Signatures = []sol.Signature{txSig}

	// Serialize and broadcast
	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("serialize transaction: %w", err)
	}

	resultSig, err := c.sc.SendTransaction(ctx, sol.Base58Data(base58.Encode(rawTx)))
	if err != nil {
		return "", fmt.Errorf("send transaction: %w", err)
	}

	return resultSig.String(), nil
}

// --- Fee Estimation ---

func (c *Client) EstimateFee(ctx context.Context, req chain.SendRequest) (string, error) {
	return "5000", nil // Solana base fee: 5000 lamports per signature
}
