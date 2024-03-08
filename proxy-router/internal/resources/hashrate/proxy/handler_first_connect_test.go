package proxy

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type getDestUserNameCase struct {
	notPreserve bool
	workerName  string
	destURL     string
	expected    string
}

var ZERO_ADDRESS = common.HexToAddress("0x0").String()

func TestGetDestUserName(t *testing.T) {
	cases := []getDestUserNameCase{
		{
			notPreserve: false,
			workerName:  "just_a_name",
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.poolWorker",
		},
		{
			notPreserve: false,
			workerName:  "accountName.",
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.",
		},
		{
			notPreserve: false,
			workerName:  "accountName.workerName",
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.workerName",
		},
		{
			notPreserve: false,
			workerName:  ".workerName",
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.workerName",
		},
		{
			notPreserve: false,
			workerName:  "accountName.workerName",
			destURL:     "stratum+tcp://.poolWorker:pwd@localhost:3333",
			expected:    ".workerName",
		},
		{
			notPreserve: false,
			workerName:  "accountName.workerName.",
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.workerName.",
		},
		{
			notPreserve: false,
			workerName:  ZERO_ADDRESS,
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.poolWorker",
		},
		{
			notPreserve: false,
			workerName:  fmt.Sprintf("accountName.%s", ZERO_ADDRESS),
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    fmt.Sprintf("poolAccount.%s", ZERO_ADDRESS),
		},
		{
			notPreserve: false,
			workerName:  "accountName.workerName",
			destURL:     "stratum+tcp://user%40lightningwallet.com:@somehost.io:4141",
			expected:    "user@lightningwallet.com",
		},
		{
			notPreserve: false,
			workerName:  "accountName.workerName",
			destURL:     "stratum+tcp://some_other_wallet_address_kind:@pplp.titan.io:4141",
			expected:    "some_other_wallet_address_kind",
		},
		{
			notPreserve: true,
			workerName:  "accountName.workerName",
			destURL:     "stratum+tcp://poolAccount.poolWorker:pwd@localhost:3333",
			expected:    "poolAccount.poolWorker",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("___%s___%s", c.workerName, c.destURL), func(t *testing.T) {
			u, err := url.Parse(c.destURL)
			require.NoError(t, err)

			require.Equal(t, c.expected, getDestUserName(c.notPreserve, c.workerName, u))
		})
	}
}
