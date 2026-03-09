package ethclient

import (
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

var publicRPCURLs = map[int][]string{
	84532: {
		"https://base-sepolia.drpc.org",
		"https://base-sepolia.gateway.tenderly.co",
		"https://sepolia.base.org",
		"https://base-sepolia-rpc.publicnode.com",
		"https://base-sepolia.therpc.io",
		"https://base-sepolia.blockpi.network/v1/rpc/public",
	},
	8453: {
		"https://base.llamarpc.com",
		"https://base-mainnet.public.blastapi.io",
		"https://base.public.blockpi.network/v1/rpc/public",
		"https://base.lava.build", 
		"https://1rpc.io/base",
		"https://base-rpc.publicnode.com",
		"https://mainnet.base.org",
	},
}

var ErrNotSupportedChain = fmt.Errorf("chain is not supported")

func GetPublicRPCURLs(chainID int) ([]string, error) {
	urls, ok := publicRPCURLs[chainID]
	if !ok {
		return nil, lib.WrapError(ErrNotSupportedChain, fmt.Errorf("chainID: %d", chainID))
	}
	return urls, nil
}
