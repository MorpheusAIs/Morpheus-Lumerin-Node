package proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	gi "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
	"go.uber.org/atomic"
)

const (
	DIAL_TIMEOUT  = 10 * time.Second
	WRITE_TIMEOUT = 10 * time.Second
)

var (
	ErrIdleWriteTimeout = fmt.Errorf("connection idle write timeout")
	ErrIdleReadTimeout  = fmt.Errorf("connection idle read timeout")
)

type StratumConnection struct {
	// config
	id               string
	address          string
	idleReadTimeout  time.Duration
	idleWriteTimeout time.Duration

	// state
	connectedAt   time.Time
	reader        *bufio.Reader
	timeoutOnce   sync.Once
	readHappened  chan struct{}
	writeHappened chan struct{}
	closedCh      chan struct{}
	closeOnce     sync.Once
	timeoutErr    atomic.Pointer[error]

	idleReadAt  atomic.Time // time when connection is going to close due to idle read (no read operation for idleReadTimeout)
	idleWriteAt atomic.Time // time when connection is going to close due to idle write (no write operation for idleWriteTimeout)
	readBuffer  []byte      // buffer for incomplete read lines
	readSync    sync.Mutex
	writeSync   sync.Mutex

	// deps
	conn net.Conn
	log  gi.ILogger
}

// CreateConnection creates a new StratumConnection and starts background timer for its closure
func CreateConnection(conn net.Conn, address string, idleReadTimeout, idleWriteTimeout time.Duration, log gi.ILogger) *StratumConnection {
	c := &StratumConnection{
		id:               address,
		address:          address,
		idleReadTimeout:  idleReadTimeout,
		idleWriteTimeout: idleWriteTimeout,

		connectedAt:   time.Now(),
		reader:        bufio.NewReader(conn),
		readHappened:  make(chan struct{}, 1),
		writeHappened: make(chan struct{}, 1),
		closedCh:      make(chan struct{}),

		conn: conn,
		log: log.With(
			"DstAddr", address,
			"DstPort", lib.ParsePort(conn.LocalAddr().String()),
		),
	}

	err := conn.SetDeadline(time.Time{})
	if err != nil {
		panic(err)
	}

	c.runTimeoutTimers()
	return c
}

// Connect connects to destination with default close timeouts
func Connect(address *url.URL, idleReadCloseTimeout, idleWriteCloseTimeout time.Duration, log gi.ILogger) (*StratumConnection, error) {
	conn, err := net.DialTimeout("tcp", address.Host, DIAL_TIMEOUT)
	if err != nil {
		return nil, err
	}

	return CreateConnection(conn, address.String(), idleReadCloseTimeout, idleWriteCloseTimeout, log), nil
}

func (c *StratumConnection) LocalPort() string {
	return lib.ParsePort(c.conn.LocalAddr().String())
}

func (c *StratumConnection) Read(ctx context.Context) (interfaces.MiningMessageGeneric, error) {
	c.readSync.Lock()
	defer c.readSync.Unlock()

	cancelRoutineDoneCh := make(chan struct{})
	defer func() {
		<-cancelRoutineDoneCh
	}()

	doneCh := make(chan struct{})
	defer close(doneCh)

	// cancellation via context is implemented using SetReadDeadline,
	// which unblocks read operation causing it to return os.ErrDeadlineExceeded
	// TODO: consider implementing it in separate goroutine instead of a goroutine per read
	go func() {
		defer close(cancelRoutineDoneCh)

		select {
		case <-ctx.Done():
			c.log.Debugf("connection %s read cancelled", c.id)
			err := c.conn.SetReadDeadline(time.Now())
			if err != nil {
				// may return ErrNetClosing if fd is already closed
				c.log.Warnf("err during setting read deadline: %s", err)
				return
			}
		case <-doneCh:
			return
		case <-c.closedCh:
			return
		}
	}()

	err := c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return nil, c.maybeTimeoutError(err)
	}

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			if len(line) > 0 {
				buf := make([]byte, len(c.readBuffer)+len(line))
				copy(buf, c.readBuffer)
				copy(buf[len(c.readBuffer):], line)
				c.readBuffer = buf
			}
			// if read was cancelled via context return context error, not deadline exceeded
			if ctx.Err() != nil && errors.Is(err, os.ErrDeadlineExceeded) {
				return nil, ctx.Err()
			}
			return nil, c.maybeTimeoutError(err)
		}

		c.readHappened <- struct{}{}
		if len(c.readBuffer) > 0 {
			newLine := make([]byte, len(c.readBuffer)+len(line))
			copy(newLine, c.readBuffer)
			copy(newLine[len(c.readBuffer):], line)
			c.readBuffer = []byte{}
			line = newLine
		}
		c.log.Debugf("<= %s", string(line))

		m, err := stratumv1_message.ParseStratumMessage(line)

		if errors.Is(err, stratumv1_message.ErrStratumV1Unknown) {
			c.log.Warnf("unknown stratum message, ignoring: %s", string(line))
			continue
		}

		if err != nil {
			err2 := fmt.Errorf("invalid stratum message: %s", string(line))
			return nil, lib.WrapError(err2, err)
		}

		return m, nil
	}
}

// Write writes message to the connection. Safe for concurrent use, cause underlying TCPConn is thread-safe
func (c *StratumConnection) Write(ctx context.Context, msg interfaces.MiningMessageGeneric) error {
	c.writeSync.Lock()
	defer c.writeSync.Unlock()

	if msg == nil {
		return fmt.Errorf("nil message write attempt")
	}

	b := append(msg.Serialize(), lib.CharNewLine)

	err := c.conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return c.maybeTimeoutError(err)
	}

	ctx, cancel := context.WithTimeout(ctx, WRITE_TIMEOUT)

	// cancellation via context is implemented using SetReadDeadline,
	// which unblocks read operation causing it to return os.ErrDeadlineExceeded
	// TODO: consider implementing it in separate goroutine instead of a goroutine per read
	cancelRoutineDoneCh := make(chan struct{})
	defer func() {
		<-cancelRoutineDoneCh
	}()

	doneCh := make(chan struct{})
	defer close(doneCh)

	go func() {
		defer cancel()
		defer close(cancelRoutineDoneCh)

		select {
		case <-ctx.Done():
			err := c.conn.SetWriteDeadline(time.Now())
			if err != nil {
				// may return ErrNetClosing if fd is already closed
				c.log.Warnf("err during setting write deadline: %s", err)
				return
			}
		case <-doneCh:
			return
		case <-c.closedCh:
			return
		}
	}()

	_, err = c.conn.Write(b)

	if err != nil {
		// if read was cancelled via context return context error, not deadline exceeded
		if ctx.Err() != nil && errors.Is(err, os.ErrDeadlineExceeded) {
			return ctx.Err()
		}
		return c.maybeTimeoutError(err)
	}

	c.writeHappened <- struct{}{}

	c.log.Debugf("=> %s", string(msg.Serialize()))

	return nil
}

func (c *StratumConnection) GetID() string {
	return c.id
}

func (c *StratumConnection) Close() error {
	err := c.conn.Close()
	if err == nil {
		c.log.Debugf("connection closed %s", c.id)
	} else {
		c.log.Warnf("connection already closed %s", c.id)
	}

	c.closeOnce.Do(func() {
		close(c.closedCh)
	})

	return err
}

func (c *StratumConnection) GetConnectedAt() time.Time {
	return c.connectedAt
}

func (c *StratumConnection) GetIdleCloseAt() time.Time {
	idleReadAt := c.idleReadAt.Load()
	idleWriteAt := c.idleWriteAt.Load()
	return minTime(idleReadAt, idleWriteAt)
}

func (c *StratumConnection) ResetIdleCloseTimers() {
	c.readHappened <- struct{}{}
	c.writeHappened <- struct{}{}
}

// runTimeoutTimers runs timers to close inactive connections. If no read or write operation
// is performed for correspondingly idleRead and idleWrite timeouts, connection will close
func (c *StratumConnection) runTimeoutTimers() {
	c.timeoutOnce.Do(func() {
		go func() {
			readTimer, writeTimer := time.NewTimer(c.idleReadTimeout), time.NewTimer(c.idleWriteTimeout)

			for {
				select {
				case <-readTimer.C:
					c.timeoutErr.Store(&ErrIdleReadTimeout)
					if !writeTimer.Stop() {
						<-writeTimer.C
					}
					c.Close()
					return
				case <-writeTimer.C:
					c.timeoutErr.Store(&ErrIdleWriteTimeout)
					if !readTimer.Stop() {
						<-readTimer.C
					}
					c.Close()
					return
				case <-c.readHappened:
					if !readTimer.Stop() {
						<-readTimer.C
					}
					readTimer.Reset(c.idleReadTimeout)
					c.idleReadAt.Store(time.Now().Add(c.idleReadTimeout))
				case <-c.writeHappened:
					if !writeTimer.Stop() {
						<-writeTimer.C
					}
					writeTimer.Reset(c.idleWriteTimeout)
					c.idleWriteAt.Store(time.Now().Add(c.idleWriteTimeout))
				case <-c.closedCh:
					if !readTimer.Stop() {
						<-readTimer.C
					}
					if !writeTimer.Stop() {
						<-writeTimer.C
					}
					return
				}
			}
		}()
	})
}

func (c *StratumConnection) maybeTimeoutError(err error) error {
	timeoutErr := c.timeoutErr.Load()
	if errors.Is(err, net.ErrClosed) && timeoutErr != nil {
		return lib.WrapError(*timeoutErr, err)
	}
	return err
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
