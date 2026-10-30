package proxyapi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

const secretPrompt = "the launch codes are 0000 and the dog is named Biscuit"

type keyPair struct {
	key  *ecdsa.PrivateKey
	priv lib.HexString
	pub  lib.HexString
	addr common.Address
}

func newKeyPair(t *testing.T) keyPair {
	t.Helper()
	k, err := crypto.GenerateKey()
	require.NoError(t, err)
	return keyPair{
		key:  k,
		priv: lib.HexString(crypto.FromECDSA(k)),
		pub:  lib.HexString(crypto.FromECDSAPub(&k.PublicKey)),
		addr: crypto.PubkeyToAddress(k.PublicKey),
	}
}

type fixedKey struct{ k lib.HexString }

func (f fixedKey) GetPrivateKey() (lib.HexString, error) { return f.k, nil }
func (f fixedKey) PrivateKeyUpdated() <-chan struct{}    { return nil }

var _ interfaces.PrKeyProvider = fixedKey{}

// ---- wire format -----------------------------------------------------------

// The whole point: with encryption on, the prompt text must not appear in the
// bytes that cross the TCP link; with the legacy encoding it does.
func TestSessionPromptRequestHidesPromptOnWire(t *testing.T) {
	rpc := m.NewMorRpc()
	consumer, provider := newKeyPair(t), newKeyPair(t)
	sessionID := common.HexToHash("0x01")
	prompt := map[string]any{"messages": []map[string]string{{"role": "user", "content": secretPrompt}}}

	plain, err := rpc.SessionPromptRequest(sessionID, prompt, provider.pub, consumer.priv, "r1", false)
	require.NoError(t, err)
	plainWire, err := json.Marshal(plain)
	require.NoError(t, err)
	require.True(t, bytes.Contains(plainWire, []byte(secretPrompt)), "legacy encoding is plaintext on the wire")

	enc, err := rpc.SessionPromptRequest(sessionID, prompt, provider.pub, consumer.priv, "r2", true)
	require.NoError(t, err)
	encWire, err := json.Marshal(enc)
	require.NoError(t, err)
	require.False(t, bytes.Contains(encWire, []byte(secretPrompt)), "encrypted prompt must not be readable on the wire")
	require.False(t, bytes.Contains(encWire, []byte("Biscuit")))

	var req m.SessionPromptReq
	require.NoError(t, json.Unmarshal(enc.Params, &req))
	require.Equal(t, m.EncECIES, req.Enc)

	// The consumer signature covers the ciphertext and the enc marker
	// (encrypt-then-sign), so the provider authenticates before decrypting.
	sig := req.Signature
	req.Signature = lib.HexString{}
	require.True(t, rpc.VerifySignature(req, sig, consumer.pub, lib.NewTestLogger()))
	req.Enc = ""
	require.False(t, rpc.VerifySignature(req, sig, consumer.pub, lib.NewTestLogger()), "stripping enc in transit must break the signature")
}

// Provider-side recovery is the exact mirror of consumer-side response
// decryption: lib.DecryptString with the receiving node's own key.
func TestOpenContentMatrix(t *testing.T) {
	rpc := m.NewMorRpc()
	consumer, provider, stranger := newKeyPair(t), newKeyPair(t), newKeyPair(t)
	prompt := map[string]string{"content": secretPrompt}
	want, _ := json.Marshal(prompt)

	enc, err := rpc.SessionPromptRequest(common.HexToHash("0x01"), prompt, provider.pub, consumer.priv, "r", true)
	require.NoError(t, err)
	var encReq m.SessionPromptReq
	require.NoError(t, json.Unmarshal(enc.Params, &encReq))

	plain, err := rpc.SessionPromptRequest(common.HexToHash("0x01"), prompt, provider.pub, consumer.priv, "r", false)
	require.NoError(t, err)
	var plainReq m.SessionPromptReq
	require.NoError(t, json.Unmarshal(plain.Params, &plainReq))
	require.Empty(t, plainReq.Enc)

	got, encrypted, err := openContent(encReq.Enc, encReq.Message, provider.priv)
	require.NoError(t, err)
	require.True(t, encrypted)
	require.JSONEq(t, string(want), string(got))

	// ciphertext for the provider is opaque to anyone else
	_, _, err = openContent(encReq.Enc, encReq.Message, stranger.priv)
	require.ErrorIs(t, err, ErrContentDecrypt)

	// unencrypted content is accepted as before
	got, encrypted, err = openContent(plainReq.Enc, plainReq.Message, provider.priv)
	require.NoError(t, err)
	require.False(t, encrypted)
	require.JSONEq(t, string(want), string(got))

	// unknown markers are never guessed at
	_, _, err = openContent("rot13", plainReq.Message, provider.priv)
	require.ErrorIs(t, err, ErrUnsupportedContentEncoding)
}

// Audio chunks carry raw bytes inside the cipher (no base64 layer); the
// legacy path keeps base64. Both must round-trip to the same bytes.
func TestStreamChunkRoundTrip(t *testing.T) {
	rpc := m.NewMorRpc()
	consumer, provider := newKeyPair(t), newKeyPair(t)
	raw := bytes.Repeat([]byte{0x00, 0xff, 0x10, 0x80}, 1024)

	enc, err := rpc.SessionPromptStreamChunkRequest(common.HexToHash("0x02"), "s1", 0, raw, provider.pub, consumer.priv, "r", true)
	require.NoError(t, err)
	var req m.SessionPromptStreamChunkReq
	require.NoError(t, json.Unmarshal(enc.Params, &req))
	require.Equal(t, m.EncECIES, req.Enc)
	got, _, err := openContent(req.Enc, req.ChunkData, provider.priv)
	require.NoError(t, err)
	require.Equal(t, raw, got)

	end, err := rpc.SessionPromptStreamEndRequest(common.HexToHash("0x02"), "s1", `{"language":"en"}`, provider.pub, consumer.priv, "r", true)
	require.NoError(t, err)
	var endReq m.SessionPromptStreamEndReq
	require.NoError(t, json.Unmarshal(end.Params, &endReq))
	got, _, err = openContent(endReq.Enc, endReq.AudioRequestParam, provider.priv)
	require.NoError(t, err)
	require.Equal(t, `{"language":"en"}`, string(got))

	tool, err := rpc.CallAgentToolRequest(common.HexToHash("0x02"), "search", map[string]any{"q": secretPrompt}, provider.pub, consumer.priv, "r", true)
	require.NoError(t, err)
	wire, _ := json.Marshal(tool)
	require.False(t, bytes.Contains(wire, []byte(secretPrompt)))
}

// A legacy plaintext request must serialize byte-for-byte as before: the new
// enc field is omitempty, so an unchanged consumer and an unchanged provider
// keep agreeing on the signed payload.
func TestLegacyPlaintextWireUnchanged(t *testing.T) {
	rpc := m.NewMorRpc()
	consumer, provider := newKeyPair(t), newKeyPair(t)
	msg, err := rpc.SessionPromptRequest(common.HexToHash("0x03"), map[string]string{"a": "b"}, provider.pub, consumer.priv, "r", false)
	require.NoError(t, err)
	require.NotContains(t, string(msg.Params), `"enc"`)

	open, err := rpc.InitiateSessionRequest(consumer.addr, provider.addr, big.NewInt(1), common.HexToHash("0x04"), consumer.priv, "r", false)
	require.NoError(t, err)
	require.NotContains(t, string(open.Params), `"enc"`)
	openEnc, err := rpc.InitiateSessionRequest(consumer.addr, provider.addr, big.NewInt(1), common.HexToHash("0x04"), consumer.priv, "r", true)
	require.NoError(t, err)
	require.Contains(t, string(openEnc.Params), `"enc":"ecies"`)
}

// ---- provider controller ---------------------------------------------------

func newEncTestController(t *testing.T, provider keyPair) *MORRPCController {
	t.Helper()
	valid, err := config.NewValidator()
	require.NoError(t, err)
	ss := storages.NewSessionStorage(storages.NewTestStorage())
	receiver := NewProxyReceiver(provider.priv, provider.pub, ss, nil, big.NewInt(1), nil, nil, nil)
	return NewMORRPCController(receiver, valid, nil, ss, provider.priv, nil)
}

func TestNetworkCapabilitiesSigned(t *testing.T) {
	rpc := m.NewMorRpc()
	consumer, provider := newKeyPair(t), newKeyPair(t)
	nonce := lib.HexString{1, 2, 3, 4, 5, 6, 7, 8}

	c := newEncTestController(t, provider)
	req, err := rpc.CapabilitiesRequest("c1", consumer.priv, nonce)
	require.NoError(t, err)

	var got *m.RpcResponse
	err = c.Handle(context.Background(), *req, lib.NewTestLogger(), func(r *m.RpcResponse) error { got = r; return nil })
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Result)

	var res m.CapabilitiesRes
	require.NoError(t, json.Unmarshal(*got.Result, &res))
	sig := res.Signature
	res.Signature = lib.HexString{}
	require.True(t, rpc.VerifySignatureAddr(res, sig, provider.addr, lib.NewTestLogger()), "capabilities are signed by the provider key")
	require.Equal(t, nonce, res.Nonce)
	require.Contains(t, res.Capabilities, m.CapPromptEncryptionECIES)
}

// ---- consumer <-> provider over real TCP -----------------------------------

// serveOnce accepts one MORRPC connection and answers it with handle, the way
// tcphandlers does (one JSON request, one JSON response).
func serveOnce(t *testing.T, handle func(ctx context.Context, msg m.RPCMessage, send SendResponse) error) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var msg m.RPCMessage
		if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&msg); err != nil {
			return
		}
		send := func(r *m.RpcResponse) error { return json.NewEncoder(conn).Encode(r) }
		if err := handle(context.Background(), msg, send); err != nil {
			_ = send(&m.RpcResponse{ID: msg.ID, Error: &m.RpcError{Code: 400, Message: err.Error(), Data: m.RPCErrorData{Timestamp: uint64(time.Now().UnixMilli()), Signature: lib.MustStringToHexString("0x0")}}})
		}
	}()
	return ln.Addr().String()
}

func newTestSender(t *testing.T, consumer keyPair) (*ProxyServiceSender, *storages.SessionStorage) {
	t.Helper()
	ss := storages.NewSessionStorage(storages.NewTestStorage())
	s := NewProxySender(big.NewInt(1), fixedKey{consumer.priv}, nil, ss, nil, time.Second, 0, 0, lib.NewTestLogger())
	return s, ss
}

func TestNegotiateAgainstUpgradedProvider(t *testing.T) {
	consumer, provider := newKeyPair(t), newKeyPair(t)
	ctrl := newEncTestController(t, provider)
	addr := serveOnce(t, func(ctx context.Context, msg m.RPCMessage, send SendResponse) error {
		return ctrl.Handle(ctx, msg, lib.NewTestLogger(), send)
	})

	sender, _ := newTestSender(t, consumer)
	caps, capsAt, encrypt := sender.negotiatePromptEncryption(context.Background(), addr, provider.addr)
	require.True(t, encrypt)
	require.NotZero(t, capsAt)
	require.Contains(t, caps, m.CapPromptEncryptionECIES)
}

// A provider whose signature does not belong to the expected address is not
// believed, whatever it advertises: the consumer sends unencrypted and does
// not cache the answer, so the next contact asks again.
func TestNegotiateRejectsForgedCapabilities(t *testing.T) {
	consumer, provider, impostor := newKeyPair(t), newKeyPair(t), newKeyPair(t)
	ctrl := newEncTestController(t, impostor)
	addr := serveOnce(t, func(ctx context.Context, msg m.RPCMessage, send SendResponse) error {
		return ctrl.Handle(ctx, msg, lib.NewTestLogger(), send)
	})

	sender, _ := newTestSender(t, consumer)
	caps, capsAt, encrypt := sender.negotiatePromptEncryption(context.Background(), addr, provider.addr)
	require.False(t, encrypt)
	require.Empty(t, caps)
	require.Zero(t, capsAt, "a forged answer is not cached")
}

// A provider on an earlier version answers "unknown method"; the consumer
// sends unencrypted, as before, and caches that answer.
func TestNegotiateAgainstLegacyProvider(t *testing.T) {
	consumer, provider := newKeyPair(t), newKeyPair(t)
	legacy := func(ctx context.Context, msg m.RPCMessage, send SendResponse) error {
		return lib.WrapError(ErrUnknownMethod, errors.New("unknown method: "+msg.Method))
	}

	sender, _ := newTestSender(t, consumer)
	caps, capsAt, encrypt := sender.negotiatePromptEncryption(context.Background(), serveOnce(t, legacy), provider.addr)
	require.False(t, encrypt)
	require.Empty(t, caps)
	require.NotZero(t, capsAt, "a legacy answer is a real answer and is cached")
}

// The cached capability set on the provider record drives the per-prompt
// decision without a network round trip while fresh.
func TestPromptEncryptionForUsesCache(t *testing.T) {
	consumer, provider := newKeyPair(t), newKeyPair(t)
	sender, ss := newTestSender(t, consumer)

	rec := &storages.User{Addr: provider.addr.Hex(), PubKey: provider.pub.Hex(), Url: "127.0.0.1:1",
		Capabilities: []string{m.CapPromptEncryptionECIES}, CapabilitiesAt: time.Now().Unix()}
	require.NoError(t, ss.AddUser(rec))
	require.True(t, sender.promptEncryptionFor(context.Background(), rec))

	none := &storages.User{Addr: provider.addr.Hex(), PubKey: provider.pub.Hex(), Url: "127.0.0.1:1",
		Capabilities: []string{}, CapabilitiesAt: time.Now().Unix()}
	require.False(t, sender.promptEncryptionFor(context.Background(), none))
}
