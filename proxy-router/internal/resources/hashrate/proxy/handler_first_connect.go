package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
	m "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
	sm "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

var (
	ErrNotStratum      = errors.New("not a stratum protocol") // means that incoming connection is not a stratum protocol
	ErrUnknownContract = errors.New("incoming connection for unknown contract")
)

type HandlerFirstConnect struct {
	// state
	handshakePipe       *pipeSync
	handshakePipeTsk    *lib.Task
	cancelHandshakePipe context.CancelFunc
	isStratum           atomic.Bool
	// deps
	proxy *Proxy
}

func NewHandlerFirstConnect(proxy *Proxy) *HandlerFirstConnect {
	return &HandlerFirstConnect{
		proxy: proxy,
	}
}

func (p *HandlerFirstConnect) Connect(ctx context.Context) error {
	p.handshakePipe = NewPipeSync(p.proxy.source, p.proxy.dest, p.handleSource, p.handleDest)
	p.handshakePipeTsk = lib.NewTask(p.handshakePipe)
	handshakeCtx, handshakeCancel := context.WithCancel(ctx)
	p.cancelHandshakePipe = handshakeCancel
	p.handshakePipeTsk.Start(handshakeCtx)
	<-p.handshakePipeTsk.Done()

	if errors.Is(p.handshakePipeTsk.Err(), context.Canceled) && ctx.Err() == nil {
		return nil
	}

	if !p.isStratum.Load() {
		return lib.WrapError(ErrNotStratum, p.handshakePipeTsk.Err())
	}

	return p.handshakePipeTsk.Err()
}

func (p *HandlerFirstConnect) handleSource(ctx context.Context, msg i.MiningMessageGeneric) (i.MiningMessageGeneric, error) {
	p.markAsStratum()

	switch msgTyped := msg.(type) {
	case *m.MiningConfigure:
		return nil, p.onMiningConfigure(ctx, msgTyped)

	case *m.MiningSubscribe:
		return nil, p.onMiningSubscribe(ctx, msgTyped)

	case *m.MiningAuthorize:
		return nil, p.onMiningAuthorize(ctx, msgTyped)

	case *m.MiningSubmit:
		return nil, fmt.Errorf("unexpected handshake message from source: %s", string(msg.Serialize()))

	default:
		p.proxy.logWarnf("unknown handshake message from source: %s", string(msg.Serialize()))
		// todo: maybe just return message, so pipe will write it
		return nil, p.proxy.dest.Write(context.Background(), msgTyped)
	}
}

func (p *HandlerFirstConnect) handleDest(ctx context.Context, msg i.MiningMessageGeneric) (i.MiningMessageGeneric, error) {
	var msgOut i.MiningMessageGeneric

	switch typed := msg.(type) {
	case *sm.MiningNotify:
		msgOut = typed

	case *sm.MiningSetDifficulty:
		msgOut = typed

	case *sm.MiningSetExtranonce:
		msgOut = nil

	case *sm.MiningSetVersionMask:
		msgOut = nil // sent manually

	// TODO: handle multiversion
	case *sm.MiningResult:
		msgOut = typed

	default:
		p.proxy.logWarnf("unknown message from dest: %s", string(typed.Serialize()))
		msgOut = typed
	}

	if msgOut != nil {
		// TODO: maybe just return message, so pipe will write it, or keep it for visibility
		return nil, p.proxy.source.Write(ctx, msgOut)
	}

	return nil, nil
}

func (p *HandlerFirstConnect) getPoolDest(contractID string) (*url.URL, error) {
	contract, isExist := p.proxy.getContractFromStoreFn(contractID)
	if !isExist {
		return nil, lib.WrapError(ErrUnknownContract, fmt.Errorf("contract id: %s", contractID))
	}
	poolDestStr := contract.PoolDest()
	return url.Parse(poolDestStr)
}

// The following handlers are responsible for managing the initial connection of the miner to the proxy and destination.
// This step requires special handling due to the "coupled" interaction between parties, unlike the destination change process,
// where the pool connection is established first, and then the miner is switched to it. This "coupled" interaction is intentionally
// designed to enable the negotiation of the version rolling mask. It's important to note that all of these handlers require
// performing reads and writes within the same goroutine. Additionally, other response handlers (identified by message ID) must be
// called right after receiving the message. This ensures that the order of messages is deterministic. If the order of messages
// during the handshake is not enforced, there is a possibility that miners may fail, for example, if the "set_version_mask"
// message is sent to the miner before receiving the "configure" result.

func (p *HandlerFirstConnect) onMiningConfigure(ctx context.Context, msgTyped *m.MiningConfigure) error {
	p.proxy.source.SetVersionRolling(msgTyped.GetVersionRolling())

	var destURL *url.URL
	contractID := msgTyped.GetLMRContractAddress()
	if contractID != "" {
		poolDestStr, err := p.getPoolDest(contractID)
		if err != nil {
			return err
		}
		destURL = poolDestStr
	}

	if destURL == nil {
		destURL = p.proxy.destURL.Load()
	}

	destConn, err := p.proxy.destFactory(ctx, destURL, p.proxy.GetSourceWorkerName(), p.proxy.source.conn.conn.RemoteAddr().String()) // set dest from contract store
	if err != nil {
		return err
	}

	p.proxy.dest = destConn
	p.handshakePipe.SetStream2(destConn)
	p.handshakePipe.StartStream2()

	p.proxy.dest.onceResult(ctx, msgTyped.GetID(), func(res *sm.MiningResult) (msg i.MiningMessageWithID, err error) {
		configureResult, err := m.ToMiningConfigureResult(res)
		if err != nil {
			p.proxy.logErrorf("expected MiningConfigureResult message, got %s", res.Serialize())
			return nil, err
		}

		vr, mask := configureResult.GetVersionRolling(), configureResult.GetVersionRollingMask()
		destConn.SetVersionRolling(vr, mask)
		p.proxy.source.SetNegotiatedVersionRollingMask(mask)

		configureResult.SetID(msgTyped.GetID())
		err = p.proxy.source.Write(ctx, configureResult)
		if err != nil {
			return nil, lib.WrapError(ErrHandshakeSource, err)
		}
		// Brains pool sends extra set-version mask message after configure result
		// In this implementation is causes issues with titan pool. Miner sends subscribe message twice
		// and then all submits become incorrect in titan. It may happen because the second message is a little bit
		// late. So we just skip it for now
		//
		// err = p.proxy.source.Write(ctx, m.NewMiningSetVersionMask(configureResult.GetVersionRollingMask()))
		// if err != nil {
		// 	return nil, lib.WrapError(ErrHandshakeSource, err)
		// }
		return nil, nil
	})

	err = p.proxy.dest.Write(ctx, msgTyped)
	if err != nil {
		return lib.WrapError(ErrHandshakeDest, err)
	}

	return nil
}

func (p *HandlerFirstConnect) onMiningSubscribe(ctx context.Context, msgTyped *m.MiningSubscribe) error {
	minerSubscribeReceived = true

	if p.proxy.dest == nil {
		destConn, err := p.proxy.destFactory(ctx, p.proxy.destURL.Load(), p.proxy.GetSourceWorkerName(), p.proxy.source.conn.conn.RemoteAddr().String())
		if err != nil {
			return err
		}

		p.proxy.dest = destConn
		p.handshakePipe.SetStream2(destConn)
		p.handshakePipe.StartStream2()
	}

	p.proxy.dest.onceResult(ctx, msgTyped.GetID(), func(res *sm.MiningResult) (msg i.MiningMessageWithID, err error) {
		subscribeResult, err := m.ToMiningSubscribeResult(res)
		if err != nil {
			return nil, fmt.Errorf("expected MiningSubscribeResult message, got %s", res.Serialize())
		}

		p.proxy.source.SetExtraNonce(subscribeResult.GetExtranonce())
		p.proxy.dest.SetExtraNonce(subscribeResult.GetExtranonce())

		subscribeResult.SetID(msgTyped.GetID())

		err = p.proxy.source.Write(ctx, subscribeResult)
		if err != nil {
			return nil, lib.WrapError(ErrHandshakeSource, err)
		}
		return nil, nil
	})

	err := p.proxy.dest.Write(ctx, msgTyped)
	if err != nil {
		return lib.WrapError(ErrHandshakeDest, err)
	}

	return nil
}

func (p *HandlerFirstConnect) onMiningAuthorize(ctx context.Context, msgTyped *m.MiningAuthorize) error {
	p.proxy.globalHashrate.OnConnect(msgTyped.GetUserName())
	p.proxy.source.SetUserName(msgTyped.GetUserName())
	p.proxy.log = p.proxy.log.With("SrcWorker", msgTyped.GetUserName())

	msgID := msgTyped.GetID()
	if !minerSubscribeReceived {
		return lib.WrapError(ErrHandshakeSource, fmt.Errorf("MiningAuthorize received before MiningSubscribe"))
	}

	msg := m.NewMiningResultSuccess(msgID)
	err := p.proxy.source.Write(ctx, msg)
	if err != nil {
		return lib.WrapError(ErrHandshakeSource, err)
	}

	var destURL *url.URL
	if p.proxy.dest == nil { // if no dest connection was established yet then use URL from proxy (default destination)
		destURL = p.proxy.destURL.Load()
	} else { // if dest connection was established during handshake
		destURL = p.proxy.dest.destUrl
	}

	if shouldPropagateWorkerName(p.proxy.notPropagateWorkerName, msgTyped.GetUserName(), destURL) {
		_, workerName, hasWorkerName := lib.SplitUsername(msgTyped.GetUserName())
		// if incoming miner was named "accountName.workerName" then we preserve worker name in destination
		if hasWorkerName {
			lib.SetWorkerName(destURL, workerName)
		}
	}

	destWorkerName := getDestUserName(p.proxy.notPropagateWorkerName, msgTyped.GetUserName(), destURL)
	p.proxy.dest.SetUserName(destWorkerName)

	// otherwise we use the same username as in source
	// this is the case for the incoming contracts,
	// where the miner userName is contractID

	userName := destURL.User.Username()
	p.proxy.dest.SetUserName(userName)
	pwd, ok := destURL.User.Password()
	if !ok {
		pwd = ""
	}
	destAuthMsg := m.NewMiningAuthorize(msgID, userName, pwd)

	p.proxy.dest.onceResult(ctx, msgID, func(res *sm.MiningResult) (msg i.MiningMessageWithID, err error) {
		if res.IsError() {
			return nil, lib.WrapError(ErrHandshakeDest, fmt.Errorf("cannot authorize in dest pool: %s", res.GetError()))
		}
		p.proxy.logInfof("handshake completed: %s", p.proxy.destURL.String())
		p.proxy.destMap.Store(p.proxy.dest)

		// close
		p.cancelHandshakePipe()

		return nil, nil
	})

	err = p.proxy.dest.Write(ctx, destAuthMsg)
	if err != nil {
		return lib.WrapError(ErrHandshakeDest, err)
	}

	return nil
}

func (p *HandlerFirstConnect) markAsStratum() {
	p.isStratum.CompareAndSwap(false, true)
}

func getDestUserName(notPreserveWorkerName bool, incomingUserName string, destURL *url.URL) string {
	if shouldPropagateWorkerName(notPreserveWorkerName, incomingUserName, destURL) {
		_, workerName, hasWorkerName := lib.SplitUsername(incomingUserName)
		// if incoming miner was named "accountName.workerName" then we preserve worker name in destination
		if hasWorkerName {
			accountName, _, _ := lib.SplitUsername(destURL.User.Username())
			return lib.JoinUsername(accountName, workerName)
		}
	}
	return destURL.User.Username()
}

// shouldPropagateWorkerName checks if the dest worker name should be set equal to the source
// Allows to track performance of each worker in the destination pool separately
// we have to exclude workerName propagation for contracts and for lightning address destination
// cause they may contain a period symbol that can be treated as a worker name separator
func shouldPropagateWorkerName(notPreserveWorkerName bool, incomingUserName string, destURL *url.URL) bool {
	if notPreserveWorkerName {
		return false
	}
	if hasLightningAddress(destURL) {
		return false
	}
	if hasPPLPHost(destURL) { // extra check for lightning pools
		return false
	}
	if isContractAddress(incomingUserName) {
		return false
	}
	return true
}

// hasLightningAddress checks if the username has lightning address (email-like, check for @ symbol)
func hasLightningAddress(url *url.URL) bool {
	return strings.Contains(url.User.Username(), "@")
}

// hasPPLPHost checks if the host name contains pplp, extra check for lightning payouts
func hasPPLPHost(url *url.URL) bool {
	return strings.Contains(strings.ToLower(url.Host), "pplp")
}

// isContractAddress checks if the username is a valid contract address, meaning it's an incoming traffic for contract
func isContractAddress(username string) bool {
	return common.IsHexAddress(username)
}
