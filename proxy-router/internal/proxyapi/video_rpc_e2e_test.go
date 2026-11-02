package proxyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// The video job RPC end to end over a real TCP connection: the consumer's
// sender signs session.video.create, the provider's controller checks it,
// calls a fake BytePlus, and the consumer verifies and decrypts the receipt
// and a relayed status.

type e2eKeyProvider struct{ key lib.HexString }

func (k e2eKeyProvider) GetPrivateKey() (lib.HexString, error) { return k.key, nil }
func (k e2eKeyProvider) PrivateKeyUpdated() <-chan struct{}    { return nil }

type e2eModelExists struct{}

func (e2eModelExists) ModelExists(context.Context, common.Hash) (bool, error) { return true, nil }

type e2eNoDial struct{}

func (e2eNoDial) TryConnect(context.Context, string) error { return nil }

// e2eServe answers one RPC per connection the way tcphandlers does.
func e2eServe(t *testing.T, c *MORRPCController) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				var msg m.RPCMessage
				if err := json.NewDecoder(conn).Decode(&msg); err != nil {
					return
				}
				err := c.Handle(context.Background(), msg, lib.NewTestLogger(), func(r *m.RpcResponse) error {
					raw, err := json.Marshal(r)
					if err != nil {
						return err
					}
					_, err = conn.Write(raw)
					return err
				})
				if err != nil {
					raw, _ := json.Marshal(&m.RpcResponse{ID: msg.ID, Error: &m.RpcError{
						Code:    400,
						Message: err.Error(),
						Data:    m.RPCErrorData{Timestamp: uint64(time.Now().UnixMilli()), Signature: lib.MustStringToHexString("0x0")},
					}})
					_, _ = conn.Write(raw)
				}
			}(conn)
		}
	}()
	return ln.Addr().String()
}

func TestVideoJobRPCEndToEnd(t *testing.T) {
	var creates atomic.Int32
	vendor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer vendor-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/contents/generations/tasks":
			creates.Add(1)
			_, _ = w.Write([]byte(`{"id":"cgt-e2e"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/contents/generations/tasks/cgt-e2e":
			_, _ = fmt.Fprintf(w, `{"id":"cgt-e2e","status":"succeeded","content":{"video_url":"https://cdn.example/clip.mp4"},"updated_at":%d}`, time.Now().Unix())
		default:
			http.NotFound(w, r)
		}
	}))
	defer vendor.Close()

	log := lib.NewTestLogger()
	providerKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	providerPrKey := lib.HexString(crypto.FromECDSA(providerKey))
	providerPub := lib.HexString(crypto.FromECDSAPub(&providerKey.PublicKey))
	providerAddr := crypto.PubkeyToAddress(providerKey.PublicKey)

	userKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	userPrKey := lib.HexString(crypto.FromECDSA(userKey))
	userPub := lib.HexString(crypto.FromECDSAPub(&userKey.PublicKey))
	userAddr := crypto.PubkeyToAddress(userKey.PublicKey)

	modelID := common.HexToHash("0x00000000000000000000000000000000000000000000000000000000000000b1")
	sessionID := common.HexToHash("0x5e55e55e55e55e55e55e55e55e55e55e55e55e55e55e55e55e55e55e55e55e5")
	session := &storages.Session{
		Id:           sessionID.Hex(),
		UserAddr:     userAddr.Hex(),
		ProviderAddr: providerAddr.Hex(),
		// One 844-second job fits; a second does not.
		EndsAt:  big.NewInt(time.Now().Add(1500 * time.Second).Unix()),
		ModelID: modelID.Hex(),
	}

	// provider
	providerSS := storages.NewSessionStorage(storages.NewTestStorage())
	require.NoError(t, providerSS.AddUserBound(&storages.User{Addr: userAddr.Hex(), PubKey: userPub.Hex()}))
	require.NoError(t, providerSS.AddSession(session))
	providerRepo := sessionrepo.NewSessionRepositoryCached(providerSS, nil, nil, log)

	valid, err := config.NewValidator()
	require.NoError(t, err)
	modelsConfig := fmt.Sprintf(`{"models":[{"modelId":%q,"modelName":"dreamina-seedance-2-5-260628","apiType":"byteplus-video","apiUrl":%q,"apiKey":"vendor-key"}]}`, modelID.Hex(), vendor.URL)
	loader := config.NewModelConfigLoader(filepath.Join(t.TempDir(), "models-config.json"), modelsConfig, valid, e2eModelExists{}, e2eNoDial{}, log)
	require.NoError(t, loader.Init())

	receiver := NewProxyReceiver(providerPrKey, providerPub, providerSS, nil, big.NewInt(1), loader, nil, providerRepo)
	controller := NewMORRPCController(receiver, valid, providerRepo, providerSS, providerPrKey, nil)
	addr := e2eServe(t, controller)

	// consumer
	consumerSS := storages.NewSessionStorage(storages.NewTestStorage())
	require.NoError(t, consumerSS.AddUser(&storages.User{Addr: providerAddr.Hex(), PubKey: providerPub.Hex(), Url: addr}))
	require.NoError(t, consumerSS.AddSession(session))
	consumerRepo := sessionrepo.NewSessionRepositoryCached(consumerSS, nil, nil, log)
	sender := NewProxySender(big.NewInt(1), e2eKeyProvider{key: userPrKey}, nil, consumerSS, consumerRepo, 90*time.Second, 3, 20, log)

	ctx := context.Background()
	five := 5
	job := &videojob.CreateRequest{Prompt: "a harbour at dusk", DurationSeconds: &five}

	receipt, err := sender.SendVideoJobCreate(ctx, sessionID, "vjob_e2e", job)
	require.NoError(t, err)
	require.Equal(t, videojob.ObjectJob, receipt.Object)
	require.Equal(t, videojob.VendorBytePlus, receipt.Vendor)
	require.Equal(t, "cgt-e2e", receipt.VendorTaskID)
	require.Equal(t, vendor.URL+"/contents/generations/tasks/cgt-e2e", receipt.StatusURL)
	require.Equal(t, "vjob_e2e", receipt.ClientJobID)
	require.Equal(t, "720p", receipt.Resolved.Resolution)
	// 5 s × 2.25 (720p) × 75 = 843.75 → 844
	require.EqualValues(t, 844, receipt.QuoteSeconds)

	// A resend of the same job gets the same receipt and no second task.
	again, err := sender.SendVideoJobCreate(ctx, sessionID, "vjob_e2e", job)
	require.NoError(t, err)
	require.Equal(t, receipt, again)
	require.EqualValues(t, 1, creates.Load())

	// A second job that no longer fits in the session is refused in
	// plaintext.
	_, err = sender.SendVideoJobCreate(ctx, sessionID, "vjob_other", job)
	require.True(t, errors.Is(err, ErrProviderRefused), "got %v", err)
	require.EqualValues(t, 1, creates.Load())

	update, err := sender.SendVideoJobStatus(ctx, sessionID, "vjob_e2e")
	require.NoError(t, err)
	require.Equal(t, videojob.StatusSucceeded, update.Status)
	require.Equal(t, "https://cdn.example/clip.mp4", update.VideoURL)
	require.NotNil(t, update.ExpiresAt)
}
