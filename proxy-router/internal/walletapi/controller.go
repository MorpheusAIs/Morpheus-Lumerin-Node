package walletapi

import (
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/gin-gonic/gin"
)

type WalletController struct {
	service interfaces.Wallet
}

func NewWalletController(service interfaces.Wallet) *WalletController {
	c := &WalletController{
		service: service,
	}

	return c
}

func (s *WalletController) RegisterRoutes(r interfaces.Router) {
	r.GET("/wallet", s.GetWallet)
	r.POST("/wallet/privateKey", s.SetupWalletPrivateKey)
	r.POST("/wallet/mnemonic", s.SetupWalletMnemonic)
	r.DELETE("/wallet", s.DeleteWallet)
}

// GetWallet godoc
//
//	@Summary		Get Wallet
//	@Description	Get wallet address
//	@Tags			wallet
//	@Produce		json
//	@Success		200	{object}	WalletRes
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
	ctx.JSON(http.StatusOK, WalletRes{Address: addr})
}

// SetupWalletPrivateKey godoc
//
//	@Summary		Setup wallet with private key
//	@Description	Setup wallet with private key
//	@Tags			wallet
//	@Produce		json
//	@Param			privatekey	body		walletapi.SetupWalletReqBody	true	"Private key"
//	@Success		200			{walletapi.WalletRes}	walletRes
//	@Router			/wallet [post]
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
//	@Param 			mnemonic	body		string	false	"Mnemonic"
//	@Param			derivationPath	body		string	false	"Derivation path"
//	@Success		200			{walletapi.WalletRes}	walletRes
//	@Router			/wallet [post]
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
//	@Success		200			{statusRes}	walletRes
//	@Router			/wallet [delete]
func (s *WalletController) DeleteWallet(ctx *gin.Context) {
	err := s.service.DeleteWallet()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, OkRes())
}
