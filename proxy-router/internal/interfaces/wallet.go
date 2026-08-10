package interfaces

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"

type Wallet interface {
	GetPrivateKey() (lib.HexString, error)
	SetPrivateKey(privateKey lib.HexString) error
	SetMnemonic(mnemonic string, derivationPath string) error
	DeleteWallet() error
	PrivateKeyUpdated() <-chan struct{}

	// GetKind reports how the key is stored ("mnemonic", "privateKey" or
	// "env") and, for a mnemonic wallet, the active derivation path. Clients
	// use this to decide whether HD account switching is available.
	GetKind() (kind string, derivationPath string, err error)

	// SetDerivationPath re-derives the active key from the already-stored
	// mnemonic. Lets a client switch HD accounts without ever handling the
	// seed phrase. Returns an error for wallets that have no mnemonic.
	SetDerivationPath(derivationPath string) error
}
