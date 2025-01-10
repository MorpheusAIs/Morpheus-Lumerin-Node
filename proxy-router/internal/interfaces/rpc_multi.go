package interfaces

type RPCEndpoints interface {
	GetURLs() []string
	SetURLs(urls []string) error
	SetURLsNoPersist(urls []string) error
	RemoveURLs() error
}
