package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/octopuswallet/octopuswallet/internal/chain"
)

type Client struct {
	name         string
	chainID      *big.Int
	nativeSymbol string
	client       *ethclient.Client
	rpcURL       string
}

func NewClient(name, rpcURL string, chainID int64, nativeSymbol string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", name, err)
	}
	return &Client{
		name:         name,
		chainID:      big.NewInt(chainID),
		nativeSymbol: nativeSymbol,
		client:       client,
		rpcURL:       rpcURL,
	}, nil
}

func (c *Client) Name() string           { return c.name }
func (c *Client) Type() chain.ChainType  { return chain.ChainTypeEVM }
func (c *Client) NativeSymbol() string    { return c.nativeSymbol }

func (c *Client) DeriveAddress(masterSeed []byte, merchantIndex, addressIndex uint32) (string, error) {
	return DeriveAddress(masterSeed, merchantIndex, addressIndex)
}

func (c *Client) DerivePrivateKey(masterSeed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	return DerivePrivateKey(masterSeed, merchantIndex, addressIndex)
}

func (c *Client) GetCurrentBlockHeight(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}

func (c *Client) ScanBlockForPayments(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]chain.IncomingTx, error) {
	block, err := c.client.BlockByNumber(ctx, new(big.Int).SetUint64(blockHeight))
	if err != nil {
		return nil, fmt.Errorf("get block %d: %w", blockHeight, err)
	}

	var txs []chain.IncomingTx

	for _, tx := range block.Transactions() {
		if tx.To() == nil {
			continue // contract creation
		}

		to := tx.To().Hex()
		if _, watched := watchAddresses[to]; !watched {
			continue
		}

		from, err := types.Sender(types.LatestSignerForChainID(c.chainID), tx)
		if err != nil {
			continue
		}

		txs = append(txs, chain.IncomingTx{
			TxHash:      tx.Hash().Hex(),
			FromAddress: from.Hex(),
			ToAddress:   to,
			Amount:      tx.Value().String(),
			Token:       "",
			BlockHeight: blockHeight,
		})
	}

	// Scan ERC-20 Transfer events
	erc20Txs, err := c.scanERC20Transfers(ctx, blockHeight, watchAddresses)
	if err == nil {
		txs = append(txs, erc20Txs...)
	}

	return txs, nil
}

func (c *Client) scanERC20Transfers(ctx context.Context, blockHeight uint64, watchAddresses map[string]struct{}) ([]chain.IncomingTx, error) {
	blockNum := new(big.Int).SetUint64(blockHeight)
	query := ethereum.FilterQuery{
		FromBlock: blockNum,
		ToBlock:   blockNum,
		Topics:    [][]common.Hash{{TransferEventSignature()}},
	}

	logs, err := c.client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	var txs []chain.IncomingTx
	for _, log := range logs {
		if len(log.Topics) < 3 {
			continue
		}
		to := common.HexToAddress(log.Topics[2].Hex())
		if _, watched := watchAddresses[to.Hex()]; !watched {
			continue
		}

		from := common.HexToAddress(log.Topics[1].Hex())
		amount := new(big.Int).SetBytes(log.Data)

		txs = append(txs, chain.IncomingTx{
			TxHash:      log.TxHash.Hex(),
			FromAddress: from.Hex(),
			ToAddress:   to.Hex(),
			Amount:      amount.String(),
			Token:       log.Address.Hex(),
			BlockHeight: blockHeight,
		})
	}
	return txs, nil
}

func (c *Client) GetTransactionConfirmations(ctx context.Context, txHash string) (uint64, error) {
	receipt, err := c.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return 0, err
	}

	currentBlock, err := c.client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	if currentBlock < receipt.BlockNumber.Uint64() {
		return 0, nil
	}
	return currentBlock - receipt.BlockNumber.Uint64() + 1, nil
}

func (c *Client) SendTransaction(ctx context.Context, req chain.SendRequest) (string, error) {
	privKey, err := crypto.ToECDSA(req.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	nonce, err := c.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("suggest gas price: %w", err)
	}

	toAddr := common.HexToAddress(req.ToAddress)
	amount := new(big.Int)
	amount.SetString(req.Amount, 10)

	var tx *types.Transaction
	if req.Token == "" {
		// Native transfer
		tx = types.NewTransaction(nonce, toAddr, amount, 21000, gasPrice, nil)
	} else {
		// ERC-20 transfer
		data, err := PackTransfer(toAddr, amount)
		if err != nil {
			return "", fmt.Errorf("pack transfer: %w", err)
		}
		tokenAddr := common.HexToAddress(req.Token)
		gasLimit, err := c.client.EstimateGas(ctx, ethereum.CallMsg{
			From: fromAddr,
			To:   &tokenAddr,
			Data: data,
		})
		if err != nil {
			gasLimit = 100000
		}
		tx = types.NewTransaction(nonce, tokenAddr, big.NewInt(0), gasLimit, gasPrice, data)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), (*ecdsa.PrivateKey)(privKey))
	if err != nil {
		return "", fmt.Errorf("sign transaction: %w", err)
	}

	if err := c.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (c *Client) EstimateFee(ctx context.Context, req chain.SendRequest) (string, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	var gasLimit uint64
	if req.Token == "" {
		gasLimit = 21000
	} else {
		gasLimit = 100000
	}

	fee := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit))
	return fee.String(), nil
}
