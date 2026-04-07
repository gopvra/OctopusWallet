package solana

import (
	"crypto/ed25519"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
)

const coinTypeSOL = 501

// DeriveAddress derives a Solana address from seed.
// Path: m/44'/501'/merchantIndex'/0/addressIndex
func DeriveAddress(seed []byte, merchantIndex, addressIndex uint32) (string, error) {
	key, err := deriveSolanaKey(seed, merchantIndex, addressIndex)
	if err != nil {
		return "", err
	}
	pubKey := ed25519.PrivateKey(key).Public().(ed25519.PublicKey)
	return base58.Encode(pubKey), nil
}

// DerivePrivateKey derives the private key bytes for a Solana address.
func DerivePrivateKey(seed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	return deriveSolanaKey(seed, merchantIndex, addressIndex)
}

func deriveSolanaKey(seed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	// Derive using BIP-32 to get deterministic entropy, then use as ed25519 seed
	key, err := wallet.DeriveKey(seed, 44, coinTypeSOL, merchantIndex, 0, addressIndex)
	if err != nil {
		return nil, fmt.Errorf("derive solana key: %w", err)
	}

	// Extract the private key bytes and use as ed25519 seed
	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("get private key: %w", err)
	}

	rawKey := privKey.Serialize()
	ed25519Seed := make([]byte, ed25519.SeedSize)
	copy(ed25519Seed, rawKey)
	return ed25519.NewKeyFromSeed(ed25519Seed), nil
}
