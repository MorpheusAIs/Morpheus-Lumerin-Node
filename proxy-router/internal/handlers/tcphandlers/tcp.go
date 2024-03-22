package tcphandlers

import (
	"context"
	"errors"
	"net"
	"net/url"
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/repositories/transport"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/allocator"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy"
)

func NewTCPHandler(
	log, connLog, proxyLog interfaces.ILogger,
	schedulerLogFactory func(contractID string) (interfaces.ILogger, error),
	notPropagateWorkerName bool, idleReadTimeout, idleWriteTimeout time.Duration,
	minerVettingShares, maxCachedDests int,
	defaultDestUrl *url.URL,
	destFactory proxy.DestConnFactory,
	hashrateFactory proxy.HashrateFactory,
	globalHashrate *hashrate.GlobalHashrate,
	hashrateCounterDefault string,
	alloc *allocator.Allocator,
	getContractFromStoreFn proxy.GetContractFromStoreFn,
) transport.Handler {
	return func(ctx context.Context, conn net.Conn) {
		addr := conn.RemoteAddr().String()
		sourceLog := connLog.Named("SRC").With("SrcAddr", addr)

		stratumConn := proxy.CreateConnection(conn, addr, idleReadTimeout, idleWriteTimeout, sourceLog)
		defer stratumConn.Close()

		sourceConn := proxy.NewSourceConn(stratumConn, sourceLog)

		schedulerLog, err := schedulerLogFactory(addr)
		defer func() {
			_ = schedulerLog.Close()
		}()

		if err != nil {
			sourceLog.Errorf("failed to create scheduler logger: %s", err)
			return
		}

		defer func() { _ = schedulerLog.Sync() }()

		url := lib.CopyURL(defaultDestUrl) // clones url
		prx := proxy.NewProxy(
			addr, sourceConn,
			destFactory, hashrateFactory,
			globalHashrate, url, notPropagateWorkerName,
			minerVettingShares, maxCachedDests,
			proxyLog.Named("PRX").With("SrcAddr", addr),
			getContractFromStoreFn,
		)
		scheduler := allocator.NewScheduler(
			prx,
			hashrateCounterDefault,
			url,
			minerVettingShares,
			hashrateFactory,
			alloc.InvokeVettedListeners,
			schedulerLog.With("SrcAddr", addr),
		)
		alloc.GetMiners().Store(scheduler)

		err = scheduler.Run(ctx)
		if err != nil {
			var logFunc func(template string, args ...interface{})
			if errors.Is(err, proxy.ErrNotStratum) {
				logFunc = connLog.Debugf
			} else if errors.Is(err, proxy.ErrUnknownContract) {
				logFunc = connLog.Warnf
			} else if errors.Is(err, context.Canceled) {
				logFunc = connLog.Infof
			} else {
				logFunc = connLog.Warnf
			}
			logFunc("proxy disconnected: %s %s", err, addr)
		}

		alloc.GetMiners().Delete(addr)
		return
	}
}
