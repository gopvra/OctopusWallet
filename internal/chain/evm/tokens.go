package evm

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ERC-20 Transfer event signature
var transferEventSig = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

// Minimal ERC-20 ABI for transfer calls
const erc20ABI = `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"},{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`

var parsedERC20ABI abi.ABI

func init() {
	var err error
	parsedERC20ABI, err = abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		panic("failed to parse ERC20 ABI: " + err.Error())
	}
}

func TransferEventSignature() common.Hash {
	return transferEventSig
}

func ERC20ABI() abi.ABI {
	return parsedERC20ABI
}

func PackTransfer(to common.Address, amount *big.Int) ([]byte, error) {
	return parsedERC20ABI.Pack("transfer", to, amount)
}

func PackBalanceOf(owner common.Address) ([]byte, error) {
	return parsedERC20ABI.Pack("balanceOf", owner)
}
