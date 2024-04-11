package morrpc

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMorRpc_generateSignature(t *testing.T) {
	m := NewMorRpc()

	params := map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}

	privateKeyHex := "3ceb688d9b87c1a468a7eadde744828ec8bb2d11c9ea52a179058e47f92f25ee"

	signature, err := m.generateSignature(params, privateKeyHex)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestMorRpc_verifySignature(t *testing.T) {
	m := NewMorRpc()

	params := map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}

	privateKeyHex := "81f44a49c40f206517efbbcca783d808914841200e0ac9a769368e1b2741e227"
	publicKey := "033e5e77f12aa67e52484ce64b64737d397098e78d54beba15a0bf6dcfdd5ae7e2"

	signature, err := m.generateSignature(params, privateKeyHex)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	publicKeyBytes, err := hex.DecodeString(publicKey)
	assert.NoError(t, err)

	paramsBytes, err := json.Marshal(params)

	isValid := m.VerifySignature(paramsBytes, signature, publicKeyBytes)
	assert.True(t, isValid)
}

func TestMorRpc_verifySignature_incorrect_params(t *testing.T) {
	m := NewMorRpc()

	params := map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}

	privateKeyHex := "81f44a49c40f206517efbbcca783d808914841200e0ac9a769368e1b2741e227"
	publicKey := "033e5e77f12aa67e52484ce64b64737d397098e78d54beba15a0bf6dcfdd5ae7e2"

	signature, err := m.generateSignature(params, privateKeyHex)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	publicKeyBytes, err := hex.DecodeString(publicKey)
	assert.NoError(t, err)

	params["param3"] = "unknown value"
	paramsBytes, err := json.Marshal(params)
	isValid := m.VerifySignature(paramsBytes, signature, publicKeyBytes)
	assert.False(t, isValid)
}
