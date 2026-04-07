package tron

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
)

const coinTypeTRON = 195

// DeriveAddress derives a TRON address from seed.
// TRON uses the same secp256k1 curve as Ethereum but with base58check encoding.
// Path: m/44'/195'/merchantIndex'/0/addressIndex
func DeriveAddress(seed []byte, merchantIndex, addressIndex uint32) (string, error) {
	key, err := wallet.DeriveKey(seed, 44, coinTypeTRON, merchantIndex, 0, addressIndex)
	if err != nil {
		return "", fmt.Errorf("derive tron key: %w", err)
	}

	privKey, err := key.ECPrivKey()
	if err != nil {
		return "", fmt.Errorf("get ec private key: %w", err)
	}

	return privKeyToTronAddress(privKey.ToECDSA()), nil
}

// DerivePrivateKey derives the private key bytes for a TRON address.
func DerivePrivateKey(seed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	key, err := wallet.DeriveKey(seed, 44, coinTypeTRON, merchantIndex, 0, addressIndex)
	if err != nil {
		return nil, fmt.Errorf("derive tron key: %w", err)
	}

	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("get ec private key: %w", err)
	}

	return crypto.FromECDSA(privKey.ToECDSA()), nil
}

// privKeyToTronAddress converts an ECDSA private key to a TRON base58check address.
func privKeyToTronAddress(priv *ecdsa.PrivateKey) string {
	// Get Ethereum-style address (keccak256 of public key, last 20 bytes)
	ethAddr := crypto.PubkeyToAddress(priv.PublicKey)

	// TRON address = 0x41 prefix + 20-byte address
	addrBytes := make([]byte, 21)
	addrBytes[0] = 0x41
	copy(addrBytes[1:], ethAddr.Bytes())

	// Base58Check encode
	return base58.CheckEncode(addrBytes[1:], addrBytes[0])
}
