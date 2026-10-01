package lib

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func PubKeyBytesToAddr(publicKey []byte) (common.Address, error) {
	pubKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return common.Address{}, fmt.Errorf("bad key: %w", err)
	}
	return crypto.PubkeyToAddress(*pubKey), nil
}

func PubKeyHexToAddr(publicKeyHex string) (common.Address, error) {
	return PubKeyBytesToAddr(common.FromHex(publicKeyHex))
}
