package lib

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDest(t *testing.T) {
	from, _ := url.Parse("tcp://shev8.192x168x1x99:anything123@stratum.slushpool.com:3333")
	SetWorkerName(from, "worker1")
	expected := "tcp://shev8.worker1:anything123@stratum.slushpool.com:3333"
	require.Equal(t, expected, from.String())
}
