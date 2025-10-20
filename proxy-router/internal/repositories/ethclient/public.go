package ethclient

import (
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

var publicRPCURLs = map[int][]string{
	421614: {
		"https://arbitrum-sepolia.blockpi.network/v1/rpc/public",
		"https://sepolia-rollup.arbitrum.io/rpc",
		"https://arbitrum-sepolia.gateway.tenderly.co",
		"https://endpoints.omniatech.io/v1/arbitrum/sepolia/public",
		"https://public.stackup.sh/api/v1/node/arbitrum-sepolia",
		"https://arbitrum-sepolia-rpc.publicnode.com",
		"https://rpc.ankr.com/arbitrum_sepolia",
		"https://arbitrum-sepolia.public.blastapi.io",
	},
	42161: {
		"https://api.zan.top/node/v1/arb/one/public",
		"https://1rpc.io/arb",
		"https://arbitrum.blockpi.network/v1/rpc/public",
		"https://arb-pokt.nodies.app",
		"https://arbitrum.drpc.org",
		"https://arbitrum.meowrpc.com",
		"https://rpc.ankr.com/arbitrum",
		"https://arbitrum-one.public.blastapi.io",
		"https://arbitrum.gateway.tenderly.co",
		"https://arbitrum-one-rpc.publicnode.com",
		"https://arbitrum-one.publicnode.com",
		"https://arb1.arbitrum.io/rpc",
		"https://arbitrum.rpc.subquery.network/public",
		"https://api.stateless.solutions/arbitrum-one/v1/demo",
		"https://public.stackup.sh/api/v1/node/arbitrum-one",
		"https://rpc.arb1.arbitrum.gateway.fm",
		"https://arb-mainnet-public.unifra.io",
		"https://arb-mainnet.g.alchemy.com/v2/demo",
	},
	84532: {
		"https://base-sepolia.drpc.org",
		"https://base-sepolia.gateway.tenderly.co",
		"https://sepolia.base.org",
		"https://base-sepolia-rpc.publicnode.com",
		"https://base-sepolia.therpc.io",
		"https://base-sepolia.blockpi.network/v1/rpc/public"
	},
	8453: {
		"https://base.llamarpc.com",
		"https://base-mainnet.public.blastapi.io",
		"https://base.public.blockpi.network/v1/rpc/public",
		"https://base.lava.build", 
		"https://1rpc.io/base",
		"https://base-rpc.publicnode.com",
		"https://mainnet.base.org"
	}
}

var ErrNotSupportedChain = fmt.Errorf("chain is not supported")

func GetPublicRPCURLs(chainID int) ([]string, error) {
	urls, ok := publicRPCURLs[chainID]
	if !ok {
		return nil, lib.WrapError(ErrNotSupportedChain, fmt.Errorf("chainID: %d", chainID))
	}
	return urls, nil
}
