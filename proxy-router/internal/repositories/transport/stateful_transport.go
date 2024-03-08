package transport

import (
	"context"
	"net"
)

// tcp, sip and quic are stateful transports
type StatefulTransportServer interface {
	Run(ctx context.Context) error
	OnConnect(Handler)
}

type StatefulTransport interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
	OnClose(func(reason error))
}

type Handler func(ctx context.Context, transport net.Conn)
