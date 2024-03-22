package proxy

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	sm "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

func TestReadCancellation(t *testing.T) {
	delay := 50 * time.Millisecond
	timeout := 1 * time.Minute
	server, client, err := lib.TCPPipe()
	require.NoError(t, err)
	defer server.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := CreateConnection(client, "", timeout, timeout, lib.NewTestLogger())

	go func() {
		// first and only write
		_, _ = server.Write(append(sm.NewMiningAuthorize(0, "0", "0").Serialize(), lib.CharNewLine))
	}()

	// read first message ok
	_, err = conn.Read(ctx)
	require.NoError(t, err)

	go func() {
		time.Sleep(delay)
		cancel()
	}()

	// read second message should block, and then be cancelled
	t1 := time.Now()
	_, err = conn.Read(ctx)

	require.ErrorIs(t, err, context.Canceled)
	require.GreaterOrEqual(t, time.Since(t1), delay)
}

func TestReadCancellation2(t *testing.T) {
	count := 10000
	client, server, err := lib.TCPPipe()
	if err != nil {
		t.Error(err)
	}
	defer server.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connClient := CreateConnection(client, "", 1*time.Minute, 1*time.Minute, &lib.LoggerMock{})
	connServer := CreateConnection(server, "", 1*time.Minute, 1*time.Minute, &lib.LoggerMock{})

	go func() {
		for i := 0; i < count; i++ {
			_ = connServer.Write(ctx, sm.NewMiningAuthorize(i, "0", "0"))
		}
	}()

	n := 0
	e := 0
	for i := 0; i < count; i++ {
		ctx, cancel := context.WithCancel(ctx)
		go cancel()
		_, err := connClient.Read(ctx)
		if err != nil {
			require.ErrorIs(t, err, context.Canceled)
			e++
		}
		n++
	}
}

func TestWriteCancellation(t *testing.T) {
	delay := 50 * time.Millisecond
	timeout := 1 * time.Minute
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stratumClient := CreateConnection(client, "", timeout, timeout, &lib.LoggerMock{})
	stratumServer := CreateConnection(server, "", timeout, timeout, &lib.LoggerMock{})

	go func() {
		// first and only read
		_, err := stratumServer.Read(context.Background())
		if err != nil {
			t.Error(err)
		}
	}()

	// write first message ok
	err := stratumClient.Write(ctx, sm.NewMiningAuthorize(0, "0", "0"))
	require.NoError(t, err)

	go func() {
		time.Sleep(delay)
		cancel()
	}()

	// write second message should block, and then be cancelled
	t1 := time.Now()
	err = stratumClient.Write(ctx, sm.NewMiningAuthorize(1, "0", "0"))

	require.ErrorIs(t, err, context.Canceled)
	require.GreaterOrEqual(t, time.Since(t1), delay)
}

func TestWriteCancellation2(t *testing.T) {
	count := 10000
	client, server, err := lib.TCPPipe()
	if err != nil {
		t.Error(err)
	}
	defer server.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connClient := CreateConnection(client, "", 1*time.Minute, 1*time.Minute, &lib.LoggerMock{})
	connServer := CreateConnection(server, "", 1*time.Minute, 1*time.Minute, &lib.LoggerMock{})

	go func() {
		for i := 0; i < count; i++ {
			msg, err := connServer.Read(ctx)
			if err != nil {
				return
			}
			require.IsType(t, &sm.MiningAuthorize{}, msg)
		}
	}()

	n := 0
	e := 0
	for i := 0; i < count; i++ {
		ctx, cancel := context.WithCancel(ctx)
		go cancel()
		err := connClient.Write(ctx, sm.NewMiningAuthorize(i, "0", "0"))

		if err != nil {
			require.ErrorIs(t, err, context.Canceled)
			e++
		}
		n++
	}
}

func TestConnTimeoutWrite(t *testing.T) {
	timeout := 50 * time.Millisecond
	timeoutLong := 1 * time.Second
	allowance := 50 * time.Millisecond
	server, client, err := lib.TCPPipe()
	require.NoError(t, err)

	ctx := context.Background()

	defer server.Close()
	defer client.Close()

	clientConn := CreateConnection(client, "", timeoutLong, timeout, lib.NewTestLogger().Named("client"))
	serverConn := CreateConnection(server, "", timeoutLong, timeoutLong, lib.NewTestLogger().Named("server"))

	go func() {
		// try to read first message
		_, _ = serverConn.Read(ctx)
		// try to read second message, will fail due to timeout
		_, _ = serverConn.Read(ctx)
	}()

	// write first message ok
	err = clientConn.Write(ctx, sm.NewMiningAuthorize(0, "0", "0"))
	require.NoError(t, err)

	// sleep to reach timeout
	time.Sleep(timeout + allowance)

	// write second message, should fail due to timeout
	err = clientConn.Write(ctx, sm.NewMiningAuthorize(0, "0", "0"))
	require.ErrorIs(t, err, ErrIdleWriteTimeout)

	// try to read message, should fail as well due to timeout
	_, err = clientConn.Read(ctx)
	require.ErrorIs(t, err, ErrIdleWriteTimeout)
}

func TestConnTimeoutRead(t *testing.T) {
	timeout := 50 * time.Millisecond
	timeoutLong := 1 * time.Second
	allowance := 50 * time.Millisecond

	server, client, err := lib.TCPPipe()
	require.NoError(t, err)
	defer server.Close()
	defer client.Close()

	ctx := context.Background()

	clientConn := CreateConnection(client, "", timeout, timeoutLong, lib.NewTestLogger().Named("client"))
	serverConn := CreateConnection(server, "", timeoutLong, timeoutLong, lib.NewTestLogger().Named("server"))

	go func() {
		_ = serverConn.Write(ctx, sm.NewMiningAuthorize(0, "0", "0"))
		_ = serverConn.Write(ctx, sm.NewMiningAuthorize(0, "0", "0"))
	}()

	// read first message ok
	_, err = clientConn.Read(ctx)
	require.NoError(t, err)

	// sleep to reach timeout
	time.Sleep(timeout + allowance)

	// read second message, should fail due to timeout
	_, err = clientConn.Read(ctx)
	require.ErrorIs(t, err, ErrIdleReadTimeout)

	// try to read message, should fail as well due to timeout
	err = clientConn.Write(ctx, sm.NewMiningAuthorize(0, "0", "0"))
	require.ErrorIs(t, err, ErrIdleReadTimeout)
}
