package ethclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/rpc"
)

// Wrapper around multiple RPC clients: round-robin entry + retries with backoff
// when endpoints return rate limits or transient errors.
type RPCClientMultiple struct {
	lock sync.RWMutex
	// rr rotates which endpoint is tried first on each top-level Call/Batch.
	rr      uint32
	clients []*rpcClient
	log     lib.ILogger
}

func NewRPCClientMultiple(urls []string, log lib.ILogger) (*RPCClientMultiple, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no RPC URLs provided")
	}

	clients := make([]*rpcClient, 0, len(urls))
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		if u == "" {
			continue
		}
		client, err := rpc.DialOptions(context.Background(), u)
		if err != nil {
			if log != nil {
				log.Warnf("skipping unreachable RPC endpoint %s: %v", u, err)
			}
			continue
		}
		clients = append(clients, &rpcClient{url: u, client: client})
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("no working RPC endpoints among %d candidates", len(urls))
	}

	return &RPCClientMultiple{clients: clients, log: log}, nil
}

func (c *RPCClientMultiple) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return c.retriableCall(ctx, func(client *rpcClient) error {
		jsonArgs, _ := json.Marshal(args)
		if c.log != nil {
			c.log.Debugf("calling %s %s", method, string(jsonArgs))
		}
		return client.client.CallContext(ctx, result, method, args...)
	})
}

func (c *RPCClientMultiple) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	return c.retriableCall(ctx, func(client *rpcClient) error {
		return client.client.BatchCallContext(ctx, b)
	})
}

func (c *RPCClientMultiple) Close() {
	for _, rpcClient := range c.getClients() {
		rpcClient.client.Close()
	}
}

func (c *RPCClientMultiple) EthSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error) {
	client := c.getClients()[0]
	return client.client.EthSubscribe(ctx, channel, args...)
}

func (c *RPCClientMultiple) GetURLs() []string {
	clients := c.getClients()
	urls := make([]string, len(clients))
	for i, rpcClient := range clients {
		urls[i] = rpcClient.url
	}
	return urls
}

func (c *RPCClientMultiple) SetURLs(urls []string) error {
	multi, err := NewRPCClientMultiple(urls, c.log)
	if err != nil {
		return err
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	for _, rpcClient := range c.clients {
		rpcClient.client.Close()
	}
	c.clients = multi.clients
	return nil
}

func (c *RPCClientMultiple) getClients() []*rpcClient {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.clients
}

func (c *RPCClientMultiple) retriableCall(ctx context.Context, fn func(client *rpcClient) error) error {
	clients := c.getClients()
	n := len(clients)
	if n == 0 {
		return fmt.Errorf("no RPC endpoints configured")
	}

	start := int(atomic.AddUint32(&c.rr, 1) % uint32(n))
	var lastErr error

	for attempt := 0; attempt < n; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		idx := (start + attempt) % n
		rpcClient := clients[idx]

		err := fn(rpcClient)
		if err == nil {
			return nil
		}

		lastErr = err
		retryable := shouldRetryRPCError(err)
		if c.log != nil {
			c.log.Debugf("RPC error (retryable=%v) endpoint=%s: %v", retryable, rpcClient.url, err)
		}
		if !retryable {
			return err
		}

		if attempt < n-1 {
			delay := backoffWithJitter(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return lastErr
}

func backoffWithJitter(attempt int) time.Duration {
	const base = 120 * time.Millisecond
	const maxDelay = 2500 * time.Millisecond

	shift := attempt
	if shift > 5 {
		shift = 5
	}
	d := base * time.Duration(1<<uint(shift))
	if d > maxDelay {
		d = maxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(d/4 + 1)))
	return d + jitter
}

func shouldRetryRPCError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var httpErr rpc.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode == 429 || httpErr.StatusCode >= 500 {
			return true
		}
		if httpErr.StatusCode == 408 || httpErr.StatusCode == 425 {
			return true
		}
		// 401/403: Cloudflare / WAF / “open in browser” pages — useless for JSON-RPC; try next endpoint
		if httpErr.StatusCode == 403 || httpErr.StatusCode == 401 {
			return true
		}
		bodyLower := strings.ToLower(string(httpErr.Body))
		if strings.Contains(bodyLower, "cloudflare") || strings.Contains(bodyLower, "<!doctype html") {
			return true
		}
		if httpErr.StatusCode >= 400 {
			return false
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if netErr, ok := urlErr.Err.(net.Error); ok && netErr.Timeout() {
			return true
		}
		return true
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "execution reverted") || strings.Contains(msg, "revert") {
		return false
	}
	if strings.Contains(msg, "429") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "over rate limit") ||
		strings.Contains(msg, "-32016") {
		return true
	}
	// Cloudflare / HTML error bodies, 403 text without typed HTTPError
	if strings.Contains(msg, "403") && (strings.Contains(msg, "forbidden") || strings.Contains(msg, "cloudflare") || strings.Contains(msg, "<!doctype")) {
		return true
	}
	if strings.Contains(msg, "just a moment") || strings.Contains(msg, "__cf_chl") || strings.Contains(msg, "cf-browser-verification") {
		return true
	}
	// Some free RPCs omit eth_call; rotate to a full node
	if strings.Contains(msg, "eth_call") && strings.Contains(msg, "not supported") {
		return true
	}
	if strings.Contains(msg, "-32601") {
		return true
	}
	if strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "tls") && strings.Contains(msg, "bad record") {
		return true
	}

	switch err.(type) {
	case *json.SyntaxError:
		return false
	case JSONError:
		return false
	default:
		return false
	}
}

type JSONError interface {
	Error() string
	ErrorCode() int
	ErrorData() interface{}
}

type rpcClient struct {
	url    string
	client *rpc.Client
}
