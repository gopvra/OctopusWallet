package evm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
)

const coinTypeETH = 60

// DeriveAddress derives an EVM address from seed.
// Path: m/44'/60'/merchantIndex'/0/addressIndex
func DeriveAddress(seed []byte, merchantIndex, addressIndex uint32) (string, error) {
	key, err := wallet.DeriveKey(seed, 44, coinTypeETH, merchantIndex, 0, addressIndex)
	if err != nil {
		return "", fmt.Errorf("derive evm key: %w", err)
	}

	privKey, err := key.ECPrivKey()
	if err != nil {
		return "", fmt.Errorf("get ec private key: %w", err)
	}

	ecdsaPrivKey := privKey.ToECDSA()
	address := crypto.PubkeyToAddress(ecdsaPrivKey.PublicKey)
	return address.Hex(), nil
}

// DerivePrivateKey derives the private key bytes for an EVM address.
func DerivePrivateKey(seed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	key, err := wallet.DeriveKey(seed, 44, coinTypeETH, merchantIndex, 0, addressIndex)
	if err != nil {
		return nil, fmt.Errorf("derive evm key: %w", err)
	}

	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("get ec private key: %w", err)
	}

	return crypto.FromECDSA(privKey.ToECDSA()), nil
}
