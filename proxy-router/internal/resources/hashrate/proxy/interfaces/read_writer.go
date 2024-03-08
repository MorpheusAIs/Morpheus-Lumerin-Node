package interfaces

import (
	"context"
	"io"
)

type StratumReadWriter interface {
	Read(ctx context.Context) (MiningMessageGeneric, error)
	Write(ctx context.Context, msg MiningMessageGeneric) error
}

type StratumReadWriteCloser interface {
	io.Closer
	StratumReadWriter
}
