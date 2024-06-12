package interfaces

type PrKeyProvider interface {
	GetPrivateKey() (string, error)
	PrivateKeyUpdated() <-chan struct{}
}
