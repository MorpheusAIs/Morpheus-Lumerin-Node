package walletapi

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type SetupWalletReqBody struct {
	PrivateKey lib.HexString `json:"privateKey" binding:"required" validate:"required,eth_addr"`
}

type WalletRes struct {
	Address common.Address `json:"address" example:"0x1234"`
}

type statusRes struct {
	Status string `json:"status" example:"ok"`
}

func OkRes() statusRes {
	return statusRes{
		Status: "ok",
	}
}
