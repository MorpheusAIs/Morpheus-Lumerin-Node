package walletapi

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type SetupWalletPrKeyReqBody struct {
	PrivateKey lib.HexString `json:"privateKey" binding:"required" validate:"required"`
}

type SetupWalletMnemonicReqBody struct {
	Mnemonic       string `json:"mnemonic" binding:"required" validate:"required"`
	DerivationPath string `json:"derivationPath" binding:"required" validate:"required"`
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
