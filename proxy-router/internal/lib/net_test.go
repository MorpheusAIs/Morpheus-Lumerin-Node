package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipe(t *testing.T) {
	client, server, err := TCPPipe()
	defer client.Close()
	defer server.Close()

	require.NoError(t, err)

	msg := []byte("hello")

	_, err = client.Write(msg)
	require.NoError(t, err)

	buf := make([]byte, len(msg))
	_, err = server.Read(buf)
	require.NoError(t, err)

	require.Equal(t, msg, buf)
}
