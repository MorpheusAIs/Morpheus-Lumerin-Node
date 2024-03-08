package proxy

import (
	"context"
	"errors"
	"fmt"

	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
	m "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/validator"
	"go.uber.org/atomic"
)

const MAX_CONSEQUENT_INVALID_SHARES = 100

type HandlerMining struct {
	// deps
	proxy                       *Proxy
	consequentInvalidShareCount atomic.Uint32
}

func NewHandlerMining(proxy *Proxy) *HandlerMining {
	return &HandlerMining{
		proxy: proxy,
	}
}

// sourceInterceptor is called when a message is received from the source after handshake
func (p *HandlerMining) sourceInterceptor(ctx context.Context, msg i.MiningMessageGeneric) (i.MiningMessageGeneric, error) {
	switch msgTyped := msg.(type) {
	case *m.MiningSubmit:
		return p.onMiningSubmit(ctx, msgTyped)
	// errors
	case *m.MiningConfigure:
		return nil, fmt.Errorf("unexpected message from source after handshake: %s", string(msg.Serialize()))
	case *m.MiningSubscribe:
		return nil, fmt.Errorf("unexpected message from source after handshake: %s", string(msg.Serialize()))
	case *m.MiningAuthorize:
		return nil, fmt.Errorf("unexpected message from source after handshake: %s", string(msg.Serialize()))
	default:
		p.proxy.logWarnf("unknown message from source: %s", string(msg.Serialize()))
		return msg, nil
	}
}

// destInterceptor is called when a message is received from the dest after handshake
func (p *HandlerMining) destInterceptor(ctx context.Context, msg i.MiningMessageGeneric) (i.MiningMessageGeneric, error) {
	switch msgTyped := msg.(type) {
	case *m.MiningSetDifficulty:
		p.proxy.logDebugf("new diff: %.0f", msgTyped.GetDifficulty())
		return msg, nil
	case *m.MiningSetVersionMask:
		p.proxy.logDebugf("got version mask: %s", msgTyped.GetVersionMask())
		return msg, nil
	case *m.MiningSetExtranonce:
		xn, xn2size := msgTyped.GetExtranonce()
		p.proxy.logDebugf("got extranonce: %s %d", xn, xn2size)
		return msg, nil
	case *m.MiningNotify:
		return msg, nil
	case *m.MiningResult:
		return msg, nil
	default:
		p.proxy.logWarnf("unknown message from dest: %s", string(msg.Serialize()))
		return msg, nil
	}
}

// onMiningSubmit is only called when handshake is completed. It doesn't require determinism
// in message ordering, so to improve performance we can use asynchronous pipe
func (p *HandlerMining) onMiningSubmit(ctx context.Context, msgTyped *m.MiningSubmit) (i.MiningMessageGeneric, error) {
	p.proxy.unansweredMsg.Add(1)

	dest := p.proxy.dest
	var res *m.MiningResult

	// searching for a job in main destination
	diff, err := dest.ValidateAndAddShare(msgTyped)
	weAccepted := err == nil

	// if share has old destination the error is job not found
	// or low difficulty (in case of job ID collision)
	if errors.Is(err, validator.ErrJobNotFound) || errors.Is(err, validator.ErrLowDifficulty) {
		// searching for a job in previous destination
		d, _, err := p.proxy.GetDestByJobIDAndValidate(msgTyped)
		if err != nil {
			weAccepted = false
			p.proxy.logWarnf("job %s not found in previous destinations", msgTyped.GetJobId())
		} else {
			weAccepted = true
			p.proxy.logWarnf("job %s found in different dest %s", msgTyped.GetJobId(), d.ID())
			dest = d
		}

	}

	if !weAccepted {
		count := p.consequentInvalidShareCount.Inc()
		if count > MAX_CONSEQUENT_INVALID_SHARES {
			p.proxy.logWarnf("too many consequent invalid shares (> %d), canceling run", MAX_CONSEQUENT_INVALID_SHARES)
			p.proxy.cancelRun()
			return nil, nil
		}
		p.proxy.source.GetStats().IncWeRejectedShares()

		if errors.Is(err, validator.ErrDuplicateShare) {
			p.proxy.logWarnf("duplicate share, jobID %s, msg id: %d", msgTyped.GetJobId(), msgTyped.GetID())
			res = m.NewMiningResultDuplicatedShare(msgTyped.GetID())
		} else if errors.Is(err, validator.ErrLowDifficulty) {
			p.proxy.logWarnf("low difficulty share jobID %s, msg id: %d, diff %.f, err %s", msgTyped.GetJobId(), msgTyped.GetID(), diff, err)
			res = m.NewMiningResultLowDifficulty(msgTyped.GetID())
		} else {
			p.proxy.logWarnf("job %s not found", msgTyped.GetJobId())
			res = m.NewMiningResultJobNotFound(msgTyped.GetID())
		}
	} else {
		p.consequentInvalidShareCount.Store(0)
		p.proxy.source.GetStats().IncWeAcceptedShares()

		// miner hashrate
		p.proxy.hashrate.OnSubmit(dest.GetDiff())
		// workername hashrate
		p.proxy.globalHashrate.OnSubmit(p.proxy.source.GetUserName(), dest.GetDiff())
		if p.proxy.hashrate.GetTotalShares() > p.proxy.vettingShares {
			select {
			case <-p.proxy.vettingDoneCh:
			default:
				close(p.proxy.vettingDoneCh)
			}
		}

		// contract hashrate
		p.proxy.onSubmitMutex.RLock()
		if p.proxy.onSubmit != nil {
			p.proxy.onSubmit(dest.GetDiff())
		}
		p.proxy.onSubmitMutex.RUnlock()

		res = m.NewMiningResultSuccess(msgTyped.GetID())
	}

	// does not wait for response from destination pool
	// TODO: implement buffering for source/dest messages
	// to avoid blocking source/dest when one of them is slow
	// and fix error handling to avoid p.cancelRun
	go func(res1 *m.MiningResult) {
		defer p.proxy.unansweredMsg.Done()

		err = p.proxy.source.Write(ctx, res1)
		if err != nil {
			p.proxy.logErrorf("cannot write response (%d) to miner: %s", res1.ID, err)
			p.proxy.cancelRun()
			return
		}

		// send and await submit response from pool
		msgTyped.SetUserName(dest.GetUserName())
		res, err := dest.WriteAwaitRes(ctx, msgTyped)
		if err != nil {
			p.proxy.logErrorf("cannot write response to pool: %s", err)
			p.proxy.cancelRun()
			return
		}

		if res.(*m.MiningResult).IsError() {
			if weAccepted {
				p.proxy.source.GetStats().IncWeAcceptedTheyRejected()
				dest.GetStats().IncWeAcceptedTheyAccepted()
				p.proxy.logWarnf("we accepted share, they rejected with err %s", res.(*m.MiningResult).GetError())
			} else {
				p.proxy.logWarnf("we rejected share, and they rejected with err %s", res.(*m.MiningResult).GetError())
			}
		} else {
			if weAccepted {
				dest.GetStats().IncWeAcceptedTheyAccepted()
			} else {
				dest.GetStats().IncWeRejectedTheyAccepted()
				p.proxy.source.GetStats().IncWeRejectedTheyAccepted()
				p.proxy.logWarnf("we rejected share, but dest accepted, diff: %.f", diff)
			}
		}
	}(res)

	return nil, nil
}
