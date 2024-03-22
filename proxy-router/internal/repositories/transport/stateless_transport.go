package transport

import "context"

// udp is a stateless transport
// TODO: update this interface
type StatelessTransportServer interface {
	Run(ctx context.Context) error
	OnConnect(func(transport StatelessTransport))
}

type StatelessTransport interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
}
