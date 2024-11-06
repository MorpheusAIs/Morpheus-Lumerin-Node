package wallet

import (
	"crypto/ecdsa"
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	mnemonic  string
	masterKey *hdkeychain.ExtendedKey
	seed      []byte
	accounts  []accounts.Account
	paths     map[common.Address]accounts.DerivationPath
}

// NewFromMnemonic returns a new wallet from a BIP-39 mnemonic.
func NewFromMnemonic(mnemonic string, passOpt ...string) (*Wallet, error) {
	seed, err := NewSeedFromMnemonic(mnemonic, passOpt...)
	if err != nil {
		return nil, err
	}

	wallet, err := newWallet(seed)
	if err != nil {
		return nil, err
	}
	wallet.mnemonic = mnemonic

	return wallet, nil
}

func newWallet(seed []byte) (*Wallet, error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		masterKey: masterKey,
		seed:      seed,
		accounts:  []accounts.Account{},
		paths:     map[common.Address]accounts.DerivationPath{},
	}, nil
}

// NewSeedFromMnemonic returns a BIP-39 seed based on a BIP-39 mnemonic.
func NewSeedFromMnemonic(mnemonic string, passOpt ...string) ([]byte, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	password := ""
	if len(passOpt) > 0 {
		password = passOpt[0]
	}

	return bip39.NewSeedWithErrorChecking(mnemonic, password)
}

// DerivePrivateKey derives the private key of the derivation path.
func (w *Wallet) DerivePrivateKey(path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	var err error
	key := w.masterKey
	for _, n := range path {
		key, err = key.Derive(n)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := key.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}

	return privateKeyECDSA, nil
}
