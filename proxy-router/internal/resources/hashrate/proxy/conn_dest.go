package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	gi "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
	sm "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/validator"
)

// ConnDest is a destination connection, a wrapper around StratumConnection,
// with destination specific state variables
type ConnDest struct {
	// config
	userName string
	destUrl  *url.URL
	destLock sync.RWMutex

	// state
	diff           atomic.Uint64
	hr             gi.Hashrate
	resultHandlers sync.Map // map[string]func(*stratumv1_message.MiningResult)

	extraNonce1     string
	extraNonce2Size int
	extraNonceLock  sync.RWMutex

	versionRolling     bool
	versionRollingMask string
	versionRollingLock sync.RWMutex

	stats     *DestStats
	validator *validator.Validator

	arDone   chan struct{}
	arCancel context.CancelFunc

	firstJobSignal chan struct{}
	firstJobOnce   sync.Once

	// deps
	conn *StratumConnection
	log  gi.ILogger
}

func NewDestConn(conn *StratumConnection, valid *validator.Validator, url *url.URL, log gi.ILogger) *ConnDest {
	return &ConnDest{
		conn:           conn,
		validator:      valid,
		userName:       url.User.Username(),
		destUrl:        url,
		stats:          &DestStats{},
		firstJobSignal: make(chan struct{}),
		log:            log,
	}
}

func ConnectDest(ctx context.Context, destURL *url.URL, valid *validator.Validator, idleReadCloseTimeout, idleWriteCloseTimeout time.Duration, log gi.ILogger) (*ConnDest, error) {
	destLog := log.Named("DST").With("DstAddr", fmt.Sprintf("%s@%s", destURL.User.Username(), destURL.Host))
	conn, err := Connect(destURL, idleReadCloseTimeout, idleWriteCloseTimeout, destLog)
	if err != nil {
		return nil, err
	}
	destLog = destLog.With("DstPort", conn.LocalPort())
	return NewDestConn(conn, valid, destURL, destLog), nil
}

func (c *ConnDest) AutoReadStart(ctx context.Context, cb func(err error)) (ok bool) {
	if c.arCancel != nil {
		return false
	}

	ctx, cancel := context.WithCancel(ctx)
	c.arCancel = cancel
	c.arDone = make(chan struct{})
	go func() {
		err := c.AutoRead(ctx)
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		cb(err)
		close(c.arDone)
	}()

	return true
}

func (c *ConnDest) AutoReadStop() error {
	if c.arCancel == nil {
		return errors.New("auto read not started")
	}
	c.arCancel()
	<-c.arDone
	c.arCancel, c.arDone = nil, nil
	return nil
}

// AutoRead reads incoming jobs from the destination connection and
// caches them so dest will not close the connection
func (c *ConnDest) AutoRead(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, err := c.Read(ctx)
		if err != nil {
			return lib.WrapError(ErrDest, err)
		}
	}
}

func (c *ConnDest) ID() string {
	return c.conn.GetID()
}

func (c *ConnDest) Read(ctx context.Context) (i.MiningMessageGeneric, error) {
	msg, err := c.conn.Read(ctx)
	if err != nil {
		return nil, lib.WrapError(ErrDest, err)
	}
	msg, err = c.readInterceptor(msg)
	if err != nil {
		return nil, lib.WrapError(ErrDest, err)
	}
	return msg, err
}

func (c *ConnDest) Write(ctx context.Context, msg i.MiningMessageGeneric) error {
	err := c.conn.Write(ctx, msg)
	if err != nil {
		return lib.WrapError(ErrDest, err)
	}
	return nil
}

func (c *ConnDest) GetExtraNonce() (extraNonce string, extraNonceSize int) {
	c.extraNonceLock.RLock()
	defer c.extraNonceLock.RUnlock()
	return c.extraNonce1, c.extraNonce2Size
}

func (c *ConnDest) SetExtraNonce(extraNonce string, extraNonceSize int) {
	c.extraNonceLock.Lock()
	defer c.extraNonceLock.Unlock()
	c.extraNonce1, c.extraNonce2Size = extraNonce, extraNonceSize
}

func (c *ConnDest) GetVersionRolling() (versionRolling bool, versionRollingMask string) {
	c.versionRollingLock.RLock()
	defer c.versionRollingLock.RUnlock()
	return c.versionRolling, c.versionRollingMask
}

func (c *ConnDest) SetVersionRolling(versionRolling bool, versionRollingMask string) {
	c.versionRollingLock.Lock()
	defer c.versionRollingLock.Unlock()
	c.versionRolling, c.versionRollingMask = versionRolling, versionRollingMask
	c.validator.SetVersionRollingMask(versionRollingMask)
}

func (c *ConnDest) GetDiff() float64 {
	return float64(c.diff.Load())
}

func (c *ConnDest) GetHR() gi.Hashrate {
	return c.hr
}

func (c *ConnDest) GetUserName() string {
	c.destLock.RLock()
	defer c.destLock.RUnlock()
	return c.userName
}

func (c *ConnDest) SetUserName(userName string) {
	c.destLock.Lock()
	defer c.destLock.Unlock()

	c.userName = userName

	newURL := lib.CopyURL(c.destUrl)
	lib.SetUserName(newURL, userName)

	c.destUrl = newURL
	c.conn.id = c.destUrl.String()
	c.conn.address = c.destUrl.String()
}

func (c *ConnDest) readInterceptor(msg i.MiningMessageGeneric) (resMsg i.MiningMessageGeneric, err error) {
	switch typed := msg.(type) {
	case *sm.MiningNotify:
		// TODO: set expiration time for all of the jobs if clean jobs flag is set to true
		xn, xnsize := c.GetExtraNonce()
		if xn == "" {
			c.log.Warn("got notify before extranonce was set")
		}
		c.validator.AddNewJob(typed, float64(c.diff.Load()), xn, xnsize)
		c.firstJobOnce.Do(func() {
			close(c.firstJobSignal)
		})
	case *sm.MiningSetDifficulty:
		c.diff.Store(uint64(typed.GetDifficulty()))
	case *sm.MiningSetExtranonce:
		c.SetExtraNonce(typed.GetExtranonce())
	case *sm.MiningSetVersionMask:
		c.SetVersionRolling(true, typed.GetVersionMask())

	// TODO: handle multiversion
	case *sm.MiningResult:
		handler, ok := c.resultHandlers.LoadAndDelete(typed.GetID())
		if ok {
			resMsg, err := handler.(ResultHandler)(typed)
			if err != nil {
				return nil, err
			}
			return resMsg, nil
		}
	}
	return msg, nil
}

// onceResult registers single time handler for the destination response with particular message ID,
// sets default timeout and does a cleanup when it expires. Returns error on result timeout
func (s *ConnDest) onceResult(ctx context.Context, msgID int, handler ResultHandler) <-chan error {
	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(ctx, RESPONSE_TIMEOUT)
	didRun := false

	s.resultHandlers.Store(msgID, func(a *sm.MiningResult) (msg i.MiningMessageWithID, err error) {
		didRun = true
		defer cancel()
		defer close(errCh)
		return handler(a)
	})

	go func() {
		<-ctx.Done()
		s.resultHandlers.Delete(msgID)
		if !didRun {
			errCh <- fmt.Errorf("dest response timeout (%s) msgID(%d)", RESPONSE_TIMEOUT, msgID)
		}
	}()

	return errCh
}

// WriteAwaitRes writes message to the destination connection and awaits for the response
func (s *ConnDest) WriteAwaitRes(parentCtx context.Context, msg i.MiningMessageWithID) (resMsg i.MiningMessageWithID, err error) {
	resCh := make(chan i.MiningMessageWithID, 1)
	msgID := msg.GetID()

	responceCtx, responceCancel := context.WithTimeout(parentCtx, RESPONSE_TIMEOUT)
	defer responceCancel()

	go func() {
		<-responceCtx.Done()
		s.resultHandlers.Delete(msgID)
	}()

	s.resultHandlers.Store(msgID, func(a *sm.MiningResult) (msg i.MiningMessageWithID, err error) {
		select {
		case <-responceCtx.Done():
		case resCh <- a:
		}
		return nil, nil
	})

	err = s.Write(parentCtx, msg)
	if err != nil {
		return nil, err
	}

	select {
	case <-responceCtx.Done():
		err := responceCtx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("dest response timeout (%s) msgID(%d)", RESPONSE_TIMEOUT, msgID)
		}
		return nil, err
	case res := <-resCh:
		return res, nil
	}
}

func (c *ConnDest) GetStats() *DestStats {
	return c.stats
}

func (c *ConnDest) HasJob(jobID string) bool {
	return c.validator.HasJob(jobID)
}

func (c *ConnDest) ValidateAndAddShare(msg *sm.MiningSubmit) (float64, error) {
	return c.validator.ValidateAndAddShare(msg)
}

func (c *ConnDest) GetLatestJob() (*validator.MiningJob, bool) {
	return c.validator.GetLatestJob()
}

func (c *ConnDest) GetFirstJobSignal() <-chan struct{} {
	return c.firstJobSignal
}

func (c *ConnDest) GetIdleCloseAt() time.Time {
	return c.conn.GetIdleCloseAt()
}

func (c *ConnDest) ResetIdleCloseTimers() {
	c.conn.ResetIdleCloseTimers()
}
