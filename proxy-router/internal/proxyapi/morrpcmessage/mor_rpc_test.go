package morrpcmesssage

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestMorRpc_generateSignature(t *testing.T) {
	m := NewMorRpc()

	params := map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}

	privateKeyHex := lib.MustStringToHexString("3ceb688d9b87c1a468a7eadde744828ec8bb2d11c9ea52a179058e47f92f25ee")

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

	privateKeyHex := lib.MustStringToHexString("81f44a49c40f206517efbbcca783d808914841200e0ac9a769368e1b2741e227")
	publicKey := lib.MustStringToHexString("033e5e77f12aa67e52484ce64b64737d397098e78d54beba15a0bf6dcfdd5ae7e2")

	signature, err := m.generateSignature(params, privateKeyHex)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	isValid := m.VerifySignature(params, signature, publicKey, nil)
	assert.True(t, isValid)
}

func TestMorRpc_verifySignature_incorrect_params(t *testing.T) {
	m := NewMorRpc()

	params := map[string]interface{}{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}

	privateKeyHex := lib.MustStringToHexString("81f44a49c40f206517efbbcca783d808914841200e0ac9a769368e1b2741e227")
	publicKey := lib.MustStringToHexString("033e5e77f12aa67e52484ce64b64737d397098e78d54beba15a0bf6dcfdd5ae7e2")

	signature, err := m.generateSignature(params, privateKeyHex)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	params["param3"] = "unknown value"
	isValid := m.VerifySignature(params, signature, publicKey, nil)
	assert.False(t, isValid)
}

func TestMorRpc_generate(t *testing.T) {
	m := NewMorRpc()

	params := map[string]interface{}{
		"user":      "2222",
		"key":       "033e5e77f12aa67e52484ce64b64737d397098e78d54beba15a0bf6dcfdd5ae7e2",
		"spend":     "10",
		"provider":  "1111",
		"timestamp": "1234567890",
	}

	privateKeyHex := lib.MustStringToHexString("81f44a49c40f206517efbbcca783d808914841200e0ac9a769368e1b2741e227")
	// publicKey := "033e5e77f12aa67e52484ce64b64737d397098e78d54beba15a0bf6dcfdd5ae7e2"

	signature, err := m.generateSignature(params, privateKeyHex)
	assert.NoError(t, err)

	hexSignature := hex.EncodeToString([]byte(signature))
	fmt.Println(hexSignature)
}

// oldPongRes mimics the PongRes struct of consumers built before the models
// field existed: unknown JSON fields are dropped on unmarshal.
type oldPongRes struct {
	Nonce     lib.HexString `json:"nonce"`
	Version   string        `json:"version,omitempty"`
	Signature lib.HexString `json:"signature"`
}

func TestPongResponceModelsExcludedFromSignature(t *testing.T) {
	prKey, err := crypto.GenerateKey()
	assert.NoError(t, err)
	prKeyBytes := crypto.FromECDSA(prKey)
	addr := crypto.PubkeyToAddress(prKey.PublicKey)

	models := []system.ModelHealthReport{
		{ModelID: "0x01", ModelType: "LLM", HasActiveBid: true, BidID: "0x02", Status: "healthy"},
	}

	rpc := NewMorRpc()
	res, err := rpc.PongResponce("1", prKeyBytes, lib.HexString{0x01, 0x02}, "v1.0.0", models)
	assert.NoError(t, err)

	// new consumer: zeroes both signature and models before verifying
	var pong PongRes
	assert.NoError(t, json.Unmarshal(*res.Result, &pong))
	assert.Len(t, pong.Models, 1)

	signature := pong.Signature
	pong.Signature = lib.HexString{}
	pong.Models = nil
	assert.True(t, rpc.VerifySignatureAddr(pong, signature, addr, lib.NewTestLogger()))

	// old consumer: models field silently dropped, verification still passes
	var old oldPongRes
	assert.NoError(t, json.Unmarshal(*res.Result, &old))
	oldSignature := old.Signature
	old.Signature = lib.HexString{}
	assert.True(t, rpc.VerifySignatureAddr(old, oldSignature, addr, lib.NewTestLogger()))
}

func TestPongResponseModelsCarryApiSpecOutsideSignature(t *testing.T) {
	prKey, err := crypto.GenerateKey()
	assert.NoError(t, err)
	prKeyBytes := crypto.FromECDSA(prKey)
	addr := crypto.PubkeyToAddress(prKey.PublicKey)

	models := []system.ModelHealthReport{{
		ModelID: "0x01", Status: system.ModelHealthStatusHealthy, LastChecked: 1,
		Api: &system.ModelApiSpec{
			Stack: "venice", ModelFamily: "qwen3",
			Thinking: &system.ThinkingSpec{Mode: system.ThinkingModeControllable},
			Bindings: map[string]*system.ParamBinding{
				system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "venice_parameters.disable_thinking", ParamType: "boolean", Value: true},
			},
		},
	}}

	rpc := NewMorRpc()
	res, err := rpc.PongResponce("req-1", prKeyBytes, lib.HexString{0x01, 0x02}, "1.0.0", models)
	assert.NoError(t, err)

	raw, err := json.Marshal(res)
	assert.NoError(t, err)
	assert.Contains(t, string(raw), `"api":{"stack":"venice"`)
	assert.Contains(t, string(raw), `"venice_parameters.disable_thinking"`)

	var pong PongRes
	assert.NoError(t, json.Unmarshal(*res.Result, &pong))
	signature := pong.Signature
	pong.Signature = lib.HexString{}
	pong.Models = nil
	assert.True(t, rpc.VerifySignatureAddr(pong, signature, addr, lib.NewTestLogger()))
}
