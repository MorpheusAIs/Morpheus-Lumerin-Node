package walletapi

import (
	"errors"
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/wallet"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
)

type WalletController struct {
	service    interfaces.Wallet
	authConfig system.HTTPAuthConfig
}

func NewWalletController(service interfaces.Wallet, authConfig system.HTTPAuthConfig) *WalletController {
	c := &WalletController{
		service:    service,
		authConfig: authConfig,
	}

	return c
}

func (s *WalletController) RegisterRoutes(r interfaces.Router) {
	r.GET("/wallet", s.authConfig.CheckAuth("get_wallet"), s.GetWallet)
	r.POST("/wallet/privateKey", s.authConfig.CheckAuth("set_wallet"), s.SetupWalletPrivateKey)
	r.POST("/wallet/mnemonic", s.authConfig.CheckAuth("set_wallet"), s.SetupWalletMnemonic)
	r.POST("/wallet/derivationPath", s.authConfig.CheckAuth("set_wallet"), s.SetDerivationPath)
	r.DELETE("/wallet", s.authConfig.CheckAuth("remove_wallet"), s.DeleteWallet)
}

// GetWallet godoc
//
//	@Summary		Get Wallet
//	@Description	Get wallet address
//	@Tags			wallet
//	@Produce		json
//	@Success		200	{object}	WalletRes
//	@Security		BasicAuth
//	@Router			/wallet [get]
func (s *WalletController) GetWallet(ctx *gin.Context) {
	prKey, err := s.service.GetPrivateKey()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kind/derivationPath are advisory: a client that cannot read them still
	// gets the address, so this stays backwards compatible.
	kind, derivationPath, kindErr := s.service.GetKind()
	if kindErr != nil {
		ctx.JSON(http.StatusOK, WalletRes{Address: addr})
		return
	}

	ctx.JSON(http.StatusOK, WalletRes{
		Address:        addr,
		Kind:           kind,
		DerivationPath: derivationPath,
	})
}

// SetDerivationPath godoc
//
//	@Summary		Switch HD account
//	@Description	Re-derives the active key from the stored mnemonic using a new derivation path. A bare index ("0", "1") is relative to m/44'/60'/0'/0. Fails if the wallet was imported as a raw private key.
//	@Tags			wallet
//	@Produce		json
//	@Param			derivationPath	body		SetDerivationPathReqBody	true	"Derivation path or account index"
//	@Success		200				{object}	WalletRes
//	@Security		BasicAuth
//	@Router			/wallet/derivationPath [post]
func (s *WalletController) SetDerivationPath(ctx *gin.Context) {
	var req SetDerivationPathReqBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.service.SetDerivationPath(req.DerivationPath); err != nil {
		// A wallet with no mnemonic is a client mistake (it should have read
		// `kind` from GET /wallet first), not a server fault.
		if errors.Is(err, wallet.ErrNoMnemonic) || errors.Is(err, wallet.ErrBadDerivationPath) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prKey, err := s.service.GetPrivateKey()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, WalletRes{
		Address:        addr,
		Kind:           wallet.WalletKindMnemonic,
		DerivationPath: req.DerivationPath,
	})
}

// SetupWalletPrivateKey godoc
//
//	@Summary		Setup wallet with private key
//	@Description	Setup wallet with private key
//	@Tags			wallet
//	@Produce		json
//	@Param			privatekey	body		SetupWalletPrKeyReqBody	true	"Private key"
//	@Success		200			{object}	WalletRes
//	@Security		BasicAuth
//	@Router			/wallet/privateKey [post]
func (s *WalletController) SetupWalletPrivateKey(ctx *gin.Context) {
	var req SetupWalletPrKeyReqBody
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.service.SetPrivateKey(req.PrivateKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prKey, err := s.service.GetPrivateKey()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, WalletRes{Address: addr})
}

// SetupWalletMnemonic godoc
//
//	@Summary		Setup wallet using mnemonic
//	@Description	Setup wallet using mnemonic
//	@Tags			wallet
//	@Produce		json
//	@Param			mnemonic	body		SetupWalletMnemonicReqBody	true	"Mnemonic"
//	@Success		200			{object}	WalletRes
//	@Security		BasicAuth
//	@Router			/wallet/mnemonic [post]
func (s *WalletController) SetupWalletMnemonic(ctx *gin.Context) {
	var req SetupWalletMnemonicReqBody
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.service.SetMnemonic(req.Mnemonic, req.DerivationPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prKey, err := s.service.GetPrivateKey()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, WalletRes{Address: addr})
}

// DeleteWallet godoc
//
//	@Summary		Remove wallet from proxy
//	@Description	Remove wallet from proxy storage
//	@Tags			wallet
//	@Produce		json
//	@Success		200	{object}	StatusRes
//	@Security		BasicAuth
//	@Router			/wallet [delete]
func (s *WalletController) DeleteWallet(ctx *gin.Context) {
	err := s.service.DeleteWallet()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, OkRes())
}
