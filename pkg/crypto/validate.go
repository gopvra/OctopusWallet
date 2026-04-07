package crypto

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// ValidateAmount checks that amount is a valid positive numeric string.
func ValidateAmount(amount string) error {
	if amount == "" {
		return fmt.Errorf("amount is required")
	}
	n := new(big.Int)
	_, ok := n.SetString(amount, 10)
	if !ok {
		return fmt.Errorf("amount must be a valid integer string")
	}
	if n.Sign() <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

// ValidateAmountOrZero allows zero value (for thresholds/limits).
func ValidateAmountOrZero(amount string) error {
	if amount == "" || amount == "0" {
		return nil
	}
	n := new(big.Int)
	_, ok := n.SetString(amount, 10)
	if !ok {
		return fmt.Errorf("must be a valid integer string")
	}
	if n.Sign() < 0 {
		return fmt.Errorf("must be non-negative")
	}
	return nil
}

var (
	evmAddressRegex    = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	btcBech32Regex     = regexp.MustCompile(`^(bc1|tb1)[a-z0-9]{25,90}$`)
	btcLegacyRegex     = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	solanaAddressRegex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
	tronAddressRegex   = regexp.MustCompile(`^T[1-9A-HJ-NP-Za-km-z]{33}$`)
)

// ValidateAddress validates an address for the given chain.
func ValidateAddress(chain, address string) error {
	if address == "" {
		return fmt.Errorf("address is required")
	}
	switch {
	case chain == "ethereum" || chain == "bsc" || chain == "polygon":
		if !evmAddressRegex.MatchString(address) {
			return fmt.Errorf("invalid EVM address format")
		}
	case chain == "bitcoin":
		if !btcBech32Regex.MatchString(address) && !btcLegacyRegex.MatchString(address) {
			return fmt.Errorf("invalid Bitcoin address format")
		}
	case chain == "solana":
		if !solanaAddressRegex.MatchString(address) {
			return fmt.Errorf("invalid Solana address format")
		}
	case chain == "tron":
		if !tronAddressRegex.MatchString(address) {
			return fmt.Errorf("invalid TRON address format")
		}
	}
	return nil
}

// ValidateHexKey validates a hex-encoded private key.
func ValidateHexKey(hexKey string) error {
	cleaned := strings.TrimPrefix(hexKey, "0x")
	_, err := hex.DecodeString(cleaned)
	if err != nil {
		return fmt.Errorf("invalid hex key: %w", err)
	}
	return nil
}
