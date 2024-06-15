package wallet

import (
	"errors"
	"sync"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/proxy-router/internal/internal/repositories/keychain"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	PRIVATE_KEY_KEY     = "private-key"
	MNEMONIC_KEY        = "mnemonic"
	DERIVATION_PATH_KEY = "mnemonic-derivation-path"
)

var (
	ErrPkey            = errors.New("cannot retrieve mnemonic or private key")
	ErrPkeyAndMnemonic = errors.New("both mnemonic and private key are stored")
)

type Wallet struct {
	storage   *keychain.Keychain
	updatedCh chan struct{}
	mutex     sync.Mutex
}

func NewWallet() *Wallet {
	return &Wallet{
		storage:   keychain.NewKeychain(),
		updatedCh: make(chan struct{}),
	}
}

// GetPrivateKey use this function to get the private key regardless of whether it was stored as a mnemonic or private key
//
// errors with ErrPkeyAndMnemonic if both mnemonic and private key are stored
func (w *Wallet) GetPrivateKey() (string, error) {
	prKey, prKeyErr := w.getStoredPrivateKey()
	mnem, derivation, mnemErr := w.getStoredMnemonic()

	if prKey != "" && mnem != "" {
		return "", errors.New("both mnemonic and private key are stored")
	}

	if prKey != "" {
		return lib.RemoveHexPrefix(prKey), prKeyErr
	}

	if mnem != "" && derivation != "" {
		wallet, err := hdwallet.NewFromMnemonic(mnem)
		if err != nil {
			return "", err
		}
		path, err := hdwallet.ParseDerivationPath(derivation)
		if err != nil {
			return "", err
		}
		account, err := wallet.Derive(path, true)
		if err != nil {
			return "", err
		}
		privateKey, err := wallet.PrivateKeyHex(account)
		if err != nil {
			return "", err
		}
		return lib.RemoveHexPrefix(privateKey), nil
	}

	var err error

	if mnemErr != nil {
		err = lib.WrapError(ErrPkey, mnemErr)
	}
	if prKeyErr != nil {
		err = lib.WrapError(err, prKeyErr)
	}

	return "", err
}

// SetPrivateKey stores the private key of the wallet
func (w *Wallet) SetPrivateKey(privateKeyOxHex string) error {
	err := w.storage.Upsert(PRIVATE_KEY_KEY, privateKeyOxHex)
	if err != nil {
		return err
	}
	// either mnemonic or private key can be stored at a time
	_, err = w.storage.Get(MNEMONIC_KEY)
	if err == nil {
		err = w.storage.Delete(MNEMONIC_KEY)
		if err != nil {
			return err
		}
	}

	_, err = w.storage.Get(DERIVATION_PATH_KEY)
	if err == nil {
		err = w.storage.Delete(DERIVATION_PATH_KEY)
		if err != nil {
			return err
		}
	}

	// notify the listeners that the private key has been updated
	close(w.updatedCh)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.updatedCh = make(chan struct{})

	return nil
}

// SetMnemonic stores the mnemonic of the wallet
func (w *Wallet) SetMnemonic(mnemonic string, derivationPath string) error {
	err := w.storage.Upsert(MNEMONIC_KEY, mnemonic)
	if err != nil {
		return err
	}
	err = w.storage.Upsert(DERIVATION_PATH_KEY, derivationPath)
	if err != nil {
		return err
	}
	// either mnemonic or private key can be stored at a time
	return w.storage.Delete(PRIVATE_KEY_KEY)
}

// getStoredPrivateKey retrieves the private key of the wallet
func (w *Wallet) getStoredPrivateKey() (string, error) {
	return w.storage.Get(PRIVATE_KEY_KEY)
}

// getStoredMnemonic retrieves the mnemonic of the wallet
func (w *Wallet) getStoredMnemonic() (string, string, error) {
	mnemonic, err := w.storage.Get(MNEMONIC_KEY)
	if err != nil {
		return "", "", err
	}

	derivationPath, err := w.storage.Get(DERIVATION_PATH_KEY)
	if err != nil {
		return "", "", err
	}

	return mnemonic, derivationPath, nil
}

func (w *Wallet) PrivateKeyUpdated() <-chan struct{} {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.updatedCh
}
