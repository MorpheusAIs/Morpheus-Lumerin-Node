package tcphandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"syscall"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type levelLog struct {
	lib.LoggerMock
	debug, warn, err int
}

func (l *levelLog) Debugf(string, ...interface{}) { l.debug++ }
func (l *levelLog) Warnf(string, ...interface{})  { l.warn++ }
func (l *levelLog) Error(...interface{})          { l.err++ }

func TestLogInboundReadErrLevel(t *testing.T) {
	reset := &net.OpError{Op: "read", Net: "tcp", Err: syscall.ECONNRESET}
	cases := []struct {
		name           string
		err            error
		debug, warn, n int
	}{
		{name: "eof", err: io.EOF, debug: 1},
		{name: "reset", err: reset, debug: 1},
		{name: "truncated json", err: io.ErrUnexpectedEOF, warn: 1},
		{name: "not json", err: &json.SyntaxError{Offset: 1}, warn: 1},
		{name: "unexpected", err: errors.New("boom"), n: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := &levelLog{}
			logInboundReadErr(log, tc.err)
			if log.debug != tc.debug || log.warn != tc.warn || log.err != tc.n {
				t.Fatalf("debug=%d warn=%d error=%d, want %d %d %d", log.debug, log.warn, log.err, tc.debug, tc.warn, tc.n)
			}
		})
	}
}

func TestGetMessageClassifiesProbe(t *testing.T) {
	t.Run("peer closes with no bytes", func(t *testing.T) {
		err := readFromPeer(t, nil)
		if !peerClosed(err) {
			t.Fatalf("err = %v, want a peer-closed error", err)
		}
	})
	t.Run("http probe", func(t *testing.T) {
		err := readFromPeer(t, []byte("GET / HTTP/1.0\r\n\r\n"))
		if !malformedMessage(err) {
			t.Fatalf("err = %v, want a malformed message", err)
		}
	})
	t.Run("truncated json", func(t *testing.T) {
		err := readFromPeer(t, []byte("{"))
		if !malformedMessage(err) {
			t.Fatalf("err = %v, want a malformed message", err)
		}
	})
}

func readFromPeer(t *testing.T, payload []byte) error {
	t.Helper()
	client, server := net.Pipe()
	t.Cleanup(func() {
		client.Close()
		server.Close()
	})
	go func() {
		if len(payload) > 0 {
			_, _ = client.Write(payload)
		}
		client.Close()
	}()
	_, err := getMessage(server)
	return err
}
