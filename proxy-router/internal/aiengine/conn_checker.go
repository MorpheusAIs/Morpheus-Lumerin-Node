package aiengine

import (
	"context"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	TimeoutConnectDefault = 3 * time.Second
)

var (
	ErrCannotParseURL = errors.New("cannot parse URL")
	ErrCannotConnect  = errors.New("cannot connect")
	ErrConnectTimeout = errors.New("connection timeout")
)

type ConnectionChecker struct{}

func (*ConnectionChecker) TryConnect(ctx context.Context, URL string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, TimeoutConnectDefault, ErrConnectTimeout)
		defer cancel()
	}
	u, err := url.Parse(URL)
	if err != nil {
		return lib.WrapError(ErrCannotParseURL, err)
	}

	var host string
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), "443")
	} else {
		host = u.Host
	}

	dialer := net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return lib.WrapError(ErrCannotConnect, err)
	}

	defer conn.Close()

	return nil
}
