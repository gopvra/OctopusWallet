package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
)

const coinTypeBTC = 0

// DeriveAddress derives a native segwit (bech32) Bitcoin address.
// Path: m/84'/0'/merchantIndex'/0/addressIndex
func DeriveAddress(seed []byte, merchantIndex, addressIndex uint32, net *chaincfg.Params) (string, error) {
	key, err := wallet.DeriveKey(seed, 84, coinTypeBTC, merchantIndex, 0, addressIndex)
	if err != nil {
		return "", fmt.Errorf("derive btc key: %w", err)
	}

	pubKey, err := key.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("get public key: %w", err)
	}

	witnessProg := btcutil.Hash160(pubKey.SerializeCompressed())
	addr, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, net)
	if err != nil {
		return "", fmt.Errorf("create witness address: %w", err)
	}

	return addr.EncodeAddress(), nil
}

// DerivePrivateKey derives the private key bytes for a Bitcoin address.
func DerivePrivateKey(seed []byte, merchantIndex, addressIndex uint32) ([]byte, error) {
	key, err := wallet.DeriveKey(seed, 84, coinTypeBTC, merchantIndex, 0, addressIndex)
	if err != nil {
		return nil, fmt.Errorf("derive btc key: %w", err)
	}

	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("get private key: %w", err)
	}

	return (*btcec.PrivateKey)(privKey).Serialize(), nil
}
