package httphandlers

type SetupWalletReqBody struct {
	PrivateKey string `json:"privateKey" binding:"required" validate:"required,eth_addr"`
}
