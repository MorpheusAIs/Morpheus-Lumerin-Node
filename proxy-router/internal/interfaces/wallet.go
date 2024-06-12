package interfaces

type Wallet interface {
	GetPrivateKey() (string, error)
	SetPrivateKey(privateKeyOxHex string) error
	SetMnemonic(mnemonic string, derivationPath string) error
	PrivateKeyUpdated() <-chan struct{}
}
