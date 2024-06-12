package wallet

import "errors"

var ErrEnvWalletSet = errors.New("cannot set private key for env wallet, switch to keychain wallet by removing WALLET_PRIVATE_KEY env var")

type EnvWallet struct {
	privKey   string
	updatedCh chan struct{}
}

func NewEnvWallet(privKey string) *EnvWallet {
	return &EnvWallet{
		updatedCh: make(chan struct{}),
		privKey:   privKey,
	}
}

func (w *EnvWallet) GetPrivateKey() (string, error) {
	return w.privKey, nil
}

func (w *EnvWallet) SetPrivateKey(privateKeyOxHex string) error {
	return ErrEnvWalletSet
}

func (w *EnvWallet) SetMnemonic(mnemonic string, derivationPath string) error {
	return ErrEnvWalletSet
}

func (w *EnvWallet) PrivateKeyUpdated() <-chan struct{} {
	return w.updatedCh
}
