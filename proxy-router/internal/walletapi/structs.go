package walletapi

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type SetupWalletPrKeyReqBody struct {
	PrivateKey lib.HexString `json:"privateKey" binding:"required" validate:"required" swaggertype:"string"`
}

type SetupWalletMnemonicReqBody struct {
	Mnemonic       string `json:"mnemonic" binding:"required" validate:"required"`
	DerivationPath string `json:"derivationPath" binding:"required" validate:"required"`
}

type WalletRes struct {
	Address common.Address `json:"address" example:"0x1234"`
	// Kind is how the key is stored: "mnemonic", "privateKey" or "env".
	// Clients use it to decide whether HD account switching is offered — a
	// privateKey or env wallet has no seed to derive further accounts from.
	Kind string `json:"kind,omitempty" example:"mnemonic"`
	// DerivationPath is the active path for a mnemonic wallet. A bare index
	// ("0", "1") is interpreted relative to m/44'/60'/0'/0.
	DerivationPath string `json:"derivationPath,omitempty" example:"0"`
}

type SetDerivationPathReqBody struct {
	DerivationPath string `json:"derivationPath" binding:"required" validate:"required"`
}

type StatusRes struct {
	Status string `json:"status" example:"ok"`
}

func OkRes() StatusRes {
	return StatusRes{
		Status: "ok",
	}
}
