package hashrate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
)

const (
	privateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
)

func TestEncryptDecrypt(t *testing.T) {
	msg := "zuck"

	privateKey2, err := crypto.HexToECDSA(privateKey)
	require.NoError(t, err)

	publicKey, ok := privateKey2.Public().(*ecdsa.PublicKey)
	require.True(t, ok)

	publicKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	hx := hex.EncodeToString(publicKeyBytes)

	encoded, err := lib.EncryptString(msg, hx)
	require.NoError(t, err)

	decoded, err := lib.DecryptString(encoded, privateKey)
	require.NoError(t, err)

	require.Equal(t, msg, decoded)
}

func TestDecryptInvalidCypher(t *testing.T) {
	terms := &EncryptedTerms{
		BaseTerms:             makeTerms(),
		ValidatorUrlEncrypted: "[object Promise]",
	}

	decrypted, err := terms.Decrypt(privateKey)
	require.ErrorIs(t, err, ErrCannotDecryptDest)

	require.Equal(t, terms.BaseTerms, decrypted.BaseTerms)
}

func TestDecryptInvalidDestURL(t *testing.T) {
	msg := "::////dd@dd@com.!;;"

	privateKey2, err := crypto.HexToECDSA(privateKey)
	require.NoError(t, err)

	publicKey, ok := privateKey2.Public().(*ecdsa.PublicKey)
	require.True(t, ok)

	publicKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	hx := hex.EncodeToString(publicKeyBytes)

	encrypted, err := lib.EncryptString(msg, hx)
	require.NoError(t, err)

	terms := &EncryptedTerms{
		BaseTerms:             makeTerms(),
		ValidatorUrlEncrypted: encrypted,
	}

	decrypted, err := terms.Decrypt(privateKey)
	require.ErrorIs(t, err, ErrInvalidDestURL)

	require.Equal(t, terms.BaseTerms, decrypted.BaseTerms)
}

func makeTerms() BaseTerms {
	return BaseTerms{
		contractID:     common.HexToAddress("0x1").Hex(),
		seller:         common.HexToAddress("0x2").Hex(),
		buyer:          common.HexToAddress("0x3").Hex(),
		startsAt:       time.Now(),
		duration:       0,
		hashrate:       0,
		price:          big.NewInt(1),
		state:          BlockchainStateRunning,
		isDeleted:      false,
		balance:        big.NewInt(1),
		hasFutureTerms: false,
		version:        0,
	}
}
