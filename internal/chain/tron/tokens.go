package tron

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/octopuswallet/octopuswallet/internal/chain"
)

// TRC-20 transfer method signature: transfer(address,uint256) = a9059cbb
const trc20TransferSig = "a9059cbb"

// parseTRC20Transfer parses a TRC-20 transfer from contract call data.
func parseTRC20Transfer(txID, contractAddr, data, fromHex string, blockHeight uint64, watchAddresses map[string]struct{}) *chain.IncomingTx {
	if len(data) < 8+64+64 { // 4 bytes selector + 32 bytes address + 32 bytes amount
		return nil
	}

	// Check if it's a transfer call
	if data[:8] != trc20TransferSig {
		return nil
	}

	// Parse 'to' address (bytes 4-36, right-padded to 32 bytes)
	toHex := "41" + data[8+24:8+64] // TRON prefix 0x41 + last 20 bytes
	toAddr := hexToTronAddress(toHex)

	if _, watched := watchAddresses[toAddr]; !watched {
		return nil
	}

	// Parse amount (bytes 36-68)
	amountHex := data[8+64 : 8+64+64]
	amount := new(big.Int)
	amount.SetString(amountHex, 16)

	fromAddr := hexToTronAddress(fromHex)
	tokenAddr := hexToTronAddress(contractAddr)

	return &chain.IncomingTx{
		TxHash:      txID,
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      amount.String(),
		Token:       tokenAddr,
		BlockHeight: blockHeight,
	}
}

// hexToTronAddress converts a hex-encoded TRON address to base58check format.
func hexToTronAddress(hexAddr string) string {
	hexAddr = strings.TrimPrefix(hexAddr, "0x")
	if len(hexAddr) == 0 {
		return ""
	}

	addrBytes, err := hex.DecodeString(hexAddr)
	if err != nil {
		return hexAddr
	}

	if len(addrBytes) < 21 {
		return hexAddr
	}

	// First byte is the version (0x41 for mainnet)
	return base58.CheckEncode(addrBytes[1:], addrBytes[0])
}

// TronAddressToHex converts a base58check TRON address to hex.
func TronAddressToHex(addr string) (string, error) {
	decoded, version, err := base58.CheckDecode(addr)
	if err != nil {
		return "", fmt.Errorf("decode tron address: %w", err)
	}
	result := make([]byte, 21)
	result[0] = version
	copy(result[1:], decoded)
	return hex.EncodeToString(result), nil
}
