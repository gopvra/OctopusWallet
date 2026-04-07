package wallet

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
)

// SeedFromMnemonic converts a BIP-39 mnemonic to a 64-byte seed.
func SeedFromMnemonic(mnemonic string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	return bip39.NewSeed(mnemonic, ""), nil
}

// GenerateMnemonic creates a new BIP-39 mnemonic (256-bit entropy, 24 words).
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// DeriveKey derives a child key from seed following BIP-44 path:
// m/purpose'/coinType'/account'/change/index
func DeriveKey(seed []byte, purpose, coinType, account, change, index uint32) (*hdkeychain.ExtendedKey, error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("create master key: %w", err)
	}

	// m/purpose'
	purposeKey, err := masterKey.Derive(hdkeychain.HardenedKeyStart + purpose)
	if err != nil {
		return nil, fmt.Errorf("derive purpose: %w", err)
	}

	// m/purpose'/coinType'
	coinTypeKey, err := purposeKey.Derive(hdkeychain.HardenedKeyStart + coinType)
	if err != nil {
		return nil, fmt.Errorf("derive coin type: %w", err)
	}

	// m/purpose'/coinType'/account'
	accountKey, err := coinTypeKey.Derive(hdkeychain.HardenedKeyStart + account)
	if err != nil {
		return nil, fmt.Errorf("derive account: %w", err)
	}

	// m/purpose'/coinType'/account'/change
	changeKey, err := accountKey.Derive(change)
	if err != nil {
		return nil, fmt.Errorf("derive change: %w", err)
	}

	// m/purpose'/coinType'/account'/change/index
	childKey, err := changeKey.Derive(index)
	if err != nil {
		return nil, fmt.Errorf("derive index: %w", err)
	}

	return childKey, nil
}
