package lib

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

var (
	ErrInvalidPrivateKey = fmt.Errorf("invalid private key")
)

func DecryptString(str string, privateKey string) (string, error) {
	pkECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", WrapError(ErrInvalidPrivateKey, err)
	}

	pkECIES := ecies.ImportECDSA(pkECDSA)
	strDecodedBytes, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}

	strDecryptedBytes, err := pkECIES.Decrypt(strDecodedBytes, nil, nil)
	if err != nil {
		return "", err
	}

	return string(strDecryptedBytes), nil
}

func EncryptString(str string, publicKeyHex string) (string, error) {
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", err
	}

	urlBytes := []byte(str)

	publicKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return "", err
	}

	pk := ecies.ImportECDSAPublic(publicKey)

	// Encrypt using ECIES
	ciphertext, err := ecies.Encrypt(rand.Reader, pk, urlBytes, nil, nil)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(ciphertext), nil
}

func PrivKeyToAddr(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}

func PrivKeyStringToAddr(privateKey string) (common.Address, error) {
	privKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return common.Address{}, WrapError(ErrInvalidPrivateKey, err)
	}

	addr, err := PrivKeyToAddr(privKey)
	if err != nil {
		return common.Address{}, WrapError(ErrInvalidPrivateKey, err)
	}
	return addr, nil
}

func MustPrivKeyToAddr(privateKey *ecdsa.PrivateKey) common.Address {
	addr, err := PrivKeyToAddr(privateKey)
	if err != nil {
		panic(err)
	}
	return addr
}

func MustPrivKeyStringToAddr(privateKey string) common.Address {
	addr, err := PrivKeyStringToAddr(privateKey)
	if err != nil {
		panic(err)
	}
	return addr
}
