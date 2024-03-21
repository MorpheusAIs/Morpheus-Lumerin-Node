package tcphandlers

import (
	"context"
	"net"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/repositories/transport"
)

func NewTCPHandler(
	log, connLog interfaces.ILogger,
	schedulerLogFactory func(contractID string) (interfaces.ILogger, error),
) transport.Handler {
	return func(ctx context.Context, conn net.Conn) {
		addr := conn.RemoteAddr().String()
		sourceLog := connLog.Named("SRC").With("SrcAddr", addr)

		schedulerLog, err := schedulerLogFactory(addr)
		defer func() {
			_ = schedulerLog.Close()
		}()

		if err != nil {
			sourceLog.Errorf("failed to create scheduler logger: %s", err)
			return
		}

		defer func() { _ = schedulerLog.Sync() }()
		return
	}
}
