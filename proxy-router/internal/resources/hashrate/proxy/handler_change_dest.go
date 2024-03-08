package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
	m "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

// HandlerChangeDest is the collection of functions that are used when the destination connection is changed
type HandlerChangeDest struct {
	proxy       *Proxy
	destFactory DestConnFactory // factory to create new destination connections
}

func NewHandlerChangeDest(proxy *Proxy, destFactory DestConnFactory) *HandlerChangeDest {
	return &HandlerChangeDest{
		proxy:       proxy,
		destFactory: destFactory,
	}
}

func (p *HandlerChangeDest) connectNewDest(ctx context.Context, newDestURL *url.URL) (*ConnDest, error) {
	newDest, err := p.destFactory(ctx, newDestURL, p.proxy.GetSourceWorkerName(), p.proxy.source.conn.conn.RemoteAddr().String())
	if err != nil {
		return nil, lib.WrapError(ErrConnectDest, err)
	}

	p.proxy.log.Debugf("new dest created")

	autoReadDone := make(chan error, 1)
	ok := newDest.AutoReadStart(ctx, func(err error) {
		if err != nil {
			p.proxy.logErrorf("error reading from new dest: %s", err)
		}
		autoReadDone <- err
		close(autoReadDone)
	})
	if !ok {
		return nil, lib.WrapError(ErrConnectDest, fmt.Errorf("autoread already started"))
	}

	p.proxy.log.Debugf("dest autoread started")

	user := newDestURL.User.Username()
	pwd, _ := newDestURL.User.Password()

	handshakeTask := lib.NewTaskFunc(func(ctx context.Context) error {
		return p.destHandshake(ctx, newDest, user, pwd)
	})

	handshakeTask.Start(ctx)

	select {
	case err := <-autoReadDone:
		handshakeTask.Stop()
		// if newDestRunTask finished first there was reading error
		// TODO: fix the case when err == nil
		return nil, lib.WrapError(ErrConnectDest, err)
	case <-handshakeTask.Done():
	}

	if handshakeTask.Err() != nil {
		return nil, lib.WrapError(ErrConnectDest, handshakeTask.Err())
	}
	p.proxy.logInfof("new destination connected url %s, localPort %s", newDestURL.String(), newDest.conn.LocalPort())

	// stops temporary reading from newDest
	err = newDest.AutoReadStop()
	if err != nil {
		return nil, err
	}
	<-autoReadDone
	p.proxy.log.Debugf("stopped new dest")
	return newDest, nil
}

// destHandshake performs handshake with the new dest when there is a dest that already connected
func (p *HandlerChangeDest) destHandshake(ctx context.Context, newDest *ConnDest, user string, pwd string) error {
	msgID := 1

	// 1. MINING.CONFIGURE
	// if miner has version mask enabled, send it to the pool
	if p.proxy.source.GetNegotiatedVersionRollingMask() != "" {
		// using the same version mask as the miner negotiated during the prev connection
		cfgMsg := m.NewMiningConfigure(msgID, nil)
		_, minBits := p.proxy.source.GetVersionRolling()
		cfgMsg.SetVersionRolling(p.proxy.source.GetNegotiatedVersionRollingMask(), minBits)
		cfgMsg.SetLMRContractAddress(user)

		res, err := newDest.WriteAwaitRes(ctx, cfgMsg)
		if err != nil {
			return lib.WrapError(ErrConnectDest, err)
		}

		cfgRes, err := m.ToMiningConfigureResult(res.(*m.MiningResult))
		if err != nil {
			return err
		}
		if cfgRes.IsError() {
			return fmt.Errorf("pool returned error: %s", cfgRes.GetError())
		}

		if cfgRes.GetVersionRollingMask() != p.proxy.source.GetNegotiatedVersionRollingMask() {
			// what to do if pool has different mask
			// TODO: consider sending set_version_mask to the pool? https://en.bitcoin.it/wiki/BIP_0310
			return fmt.Errorf("pool returned different version rolling mask: %s", cfgRes.GetVersionRollingMask())
		}

		newDest.SetVersionRolling(true, cfgRes.GetVersionRollingMask())
		p.proxy.log.Debugf("configure result received")
	}

	// 2. MINING.SUBSCRIBE
	msgID++
	gotResultCh := make(chan struct{})
	newDest.onceResult(ctx, msgID, func(a *m.MiningResult) (msg i.MiningMessageWithID, err error) {
		subRes, err := m.ToMiningSubscribeResult(a)
		if err != nil {
			return nil, err
		}
		if subRes.IsError() {
			return nil, fmt.Errorf("pool returned error: %s", subRes.GetError())
		}

		newDest.SetExtraNonce(subRes.GetExtranonce())
		p.proxy.log.Debugf("subscribe result received")
		close(gotResultCh)
		return nil, nil
	})

	err := newDest.Write(ctx, m.NewMiningSubscribe(msgID, "stratum-proxy", "1.0.0"))
	if err != nil {
		return lib.WrapError(ErrConnectDest, err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-gotResultCh:
	}

	// 3. MINING.AUTHORIZE
	msgID++

	res, err := newDest.WriteAwaitRes(ctx, m.NewMiningAuthorize(msgID, user, pwd))
	if err != nil {
		return lib.WrapError(ErrConnectDest, err)
	}

	authRes := res.(*m.MiningResult)
	if authRes.IsError() {
		return lib.WrapError(ErrConnectDest, lib.WrapError(ErrNotAuthorizedPool, fmt.Errorf("%s", authRes.GetError())))
	}

	// we need to get a job from the pool before we stop reading
	// so we use it during handshake
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-newDest.GetFirstJobSignal():
	}

	p.proxy.log.Debugf("authorize success")
	return nil
}

func (p *HandlerChangeDest) resendRelevantNotifications(ctx context.Context, newDest *ConnDest) error {
	// resend relevant notifications to the miner
	// 1. SET_VERSION_MASK
	_, versionMask := newDest.GetVersionRolling()
	err := p.proxy.source.Write(ctx, m.NewMiningSetVersionMask(versionMask))
	if err != nil {
		return lib.WrapError(ErrChangeDest, err)
	}
	p.proxy.log.Debugf("set version mask sent")

	job, ok := newDest.GetLatestJob()
	if !ok {
		return lib.WrapError(ErrChangeDest, errors.New("no job available"))
	}

	// 2. SET_EXTRANONCE
	err = p.proxy.source.Write(ctx, m.NewMiningSetExtranonce(job.GetExtraNonce1(), job.GetExtraNonce2Size()))
	if err != nil {
		return lib.WrapError(ErrChangeDest, err)
	}
	p.proxy.source.SetExtraNonce(job.GetExtraNonce1(), job.GetExtraNonce2Size())
	p.proxy.log.Debugf("extranonce sent")

	// 3. SET_DIFFICULTY
	err = p.proxy.source.Write(ctx, m.NewMiningSetDifficulty(job.GetDiff()))
	if err != nil {
		return lib.WrapError(ErrChangeDest, err)
	}
	p.proxy.log.Debugf("set difficulty sent")

	// 4. NOTIFY
	msg := job.GetNotify()
	msg.SetCleanJobs(true)

	err = p.proxy.source.Write(ctx, msg)
	if err != nil {
		return lib.WrapError(ErrChangeDest, err)
	}
	p.proxy.log.Debugf("notify sent")

	return nil
}
