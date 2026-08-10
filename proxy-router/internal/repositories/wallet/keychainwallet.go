package wallet

import (
	"errors"
	"sync"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/keychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	PRIVATE_KEY_KEY     = "private-key"
	MNEMONIC_KEY        = "mnemonic"
	DERIVATION_PATH_KEY = "mnemonic-derivation-path"
)

// How the active wallet's key material is stored.
const (
	WalletKindMnemonic   = "mnemonic"
	WalletKindPrivateKey = "privateKey"
	WalletKindEnv        = "env"
)

var (
	ErrWalletNotSet      = errors.New("wallet not set")
	ErrWallet            = errors.New("cannot retrieve mnemonic or private key")
	ErrNoMnemonic        = errors.New("wallet has no stored mnemonic; HD accounts require a mnemonic wallet")
	ErrBadDerivationPath = errors.New("invalid derivation path")
)

type KeychainWallet struct {
	storage   i.KeyValueStorage
	updatedCh chan struct{}
	mutex     sync.Mutex
}

func NewKeychainWallet(keychain i.KeyValueStorage) *KeychainWallet {
	return &KeychainWallet{
		storage:   keychain,
		updatedCh: make(chan struct{}),
	}
}

// GetPrivateKey use this function to get the private key regardless of whether it was stored as a mnemonic or private key
//
// errors with ErrPkeyAndMnemonic if both mnemonic and private key are stored
func (w *KeychainWallet) GetPrivateKey() (lib.HexString, error) {
	prKey, prKeyErr := w.getStoredPrivateKey()
	mnem, derivation, mnemErr := w.getStoredMnemonic()

	if errors.Is(prKeyErr, keychain.ErrKeyNotFound) && errors.Is(mnemErr, keychain.ErrKeyNotFound) {
		return nil, ErrWalletNotSet
	}

	if prKey != nil && mnem != "" {
		return nil, errors.New("both mnemonic and private key are stored")
	}

	if prKey != nil {
		return prKey, nil
	}

	if mnem != "" && derivation != "" {
		return w.mnemonicToPrivateKey(mnem, derivation)
	}

	var err = ErrWallet

	if mnemErr != nil && !errors.Is(mnemErr, keychain.ErrKeyNotFound) {
		err = lib.WrapError(err, mnemErr)
	}
	if prKeyErr != nil && !errors.Is(mnemErr, keychain.ErrKeyNotFound) {
		err = lib.WrapError(err, prKeyErr)
	}

	return nil, err
}

// SetPrivateKey stores the private key of the wallet
func (w *KeychainWallet) SetPrivateKey(privateKey lib.HexString) error {
	err := w.storage.Upsert(PRIVATE_KEY_KEY, privateKey.Hex())
	if err != nil {
		return err
	}
	// either mnemonic or private key can be stored at a time
	err = w.storage.DeleteIfExists(MNEMONIC_KEY)
	if err != nil {
		return err
	}

	err = w.storage.DeleteIfExists(DERIVATION_PATH_KEY)
	if err != nil {
		return err
	}

	// notify the listeners that the private key has been updated
	w.notifyUpdated()

	return nil
}

// SetMnemonic stores the mnemonic of the wallet
func (w *KeychainWallet) SetMnemonic(mnemonic string, derivationPath string) error {
	err := w.storage.Upsert(MNEMONIC_KEY, mnemonic)
	if err != nil {
		return err
	}
	err = w.storage.Upsert(DERIVATION_PATH_KEY, derivationPath)
	if err != nil {
		return err
	}

	// either mnemonic or private key can be stored at a time
	err = w.storage.DeleteIfExists(PRIVATE_KEY_KEY)
	if err != nil {
		return err
	}

	w.notifyUpdated()

	return nil
}

func (w *KeychainWallet) DeleteWallet() error {
	err := w.storage.DeleteIfExists(PRIVATE_KEY_KEY)
	if err != nil {
		return err
	}

	err = w.storage.DeleteIfExists(MNEMONIC_KEY)
	if err != nil {
		return err
	}

	err = w.storage.DeleteIfExists(DERIVATION_PATH_KEY)
	if err != nil {
		return err
	}

	w.notifyUpdated()

	return nil
}

// GetKind reports how the active wallet is stored, and — for a mnemonic
// wallet — which derivation path is currently selected.
//
// The desktop app needs this to know whether HD account switching is even
// possible: a wallet imported as a raw private key has no seed to derive
// further accounts from.
func (w *KeychainWallet) GetKind() (kind string, derivationPath string, err error) {
	prKey, prKeyErr := w.getStoredPrivateKey()
	mnem, derivation, mnemErr := w.getStoredMnemonic()

	if errors.Is(prKeyErr, keychain.ErrKeyNotFound) && errors.Is(mnemErr, keychain.ErrKeyNotFound) {
		return "", "", ErrWalletNotSet
	}
	if mnem != "" {
		return WalletKindMnemonic, derivation, nil
	}
	if prKey != nil {
		return WalletKindPrivateKey, "", nil
	}
	return "", "", ErrWallet
}

// SetDerivationPath re-derives the active key from the ALREADY STORED mnemonic.
//
// This exists so the desktop app can switch between HD accounts without ever
// holding the seed phrase itself. The app discards the mnemonic after
// onboarding (it only ever POSTs it), and re-prompting the user for it on every
// account switch would be both hostile and a good way to train people to type
// their seed into things.
//
// Errors with ErrNoMnemonic if the wallet was imported as a raw private key.
func (w *KeychainWallet) SetDerivationPath(derivationPath string) error {
	mnem, _, err := w.getStoredMnemonic()
	if err != nil {
		if errors.Is(err, keychain.ErrKeyNotFound) {
			return ErrNoMnemonic
		}
		return err
	}
	if mnem == "" {
		return ErrNoMnemonic
	}

	// Validate before persisting, so a bad path cannot leave the wallet
	// pointing at something underivable.
	if _, err := w.mnemonicToPrivateKey(mnem, derivationPath); err != nil {
		return lib.WrapError(ErrBadDerivationPath, err)
	}

	if err := w.storage.Upsert(DERIVATION_PATH_KEY, derivationPath); err != nil {
		return err
	}

	w.notifyUpdated()
	return nil
}

// getStoredPrivateKey retrieves the private key of the wallet
func (w *KeychainWallet) getStoredPrivateKey() (lib.HexString, error) {
	prKey, err := w.storage.Get(PRIVATE_KEY_KEY)
	if err != nil {
		return nil, err
	}
	return lib.StringToHexString(prKey)
}

// getStoredMnemonic retrieves the mnemonic of the wallet
func (w *KeychainWallet) getStoredMnemonic() (string, string, error) {
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

func (w *KeychainWallet) mnemonicToPrivateKey(mnemonic, derivationPath string) (lib.HexString, error) {
	wallet, err := NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	path, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, err
	}
	prKey, err := wallet.DerivePrivateKey(path)
	if err != nil {
		return nil, err
	}
	return crypto.FromECDSA(prKey), nil
}

func (w *KeychainWallet) PrivateKeyUpdated() <-chan struct{} {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.updatedCh
}

func (w *KeychainWallet) notifyUpdated() {
	close(w.updatedCh)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.updatedCh = make(chan struct{})
}
