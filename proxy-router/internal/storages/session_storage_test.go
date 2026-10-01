package storages

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestAddSession(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	session := &Session{
		Id:               "0x0",
		UserAddr:         "0x1",
		ProviderAddr:     "0x2",
		EndsAt:           big.NewInt(100),
		ModelID:          "0x3",
		TPSScaled1000Arr: []int{1, 2, 3},
		TTFTMsArr:        []int{4, 5, 6},
		FailoverEnabled:  true,
	}

	err := sessionStorage.AddSession(session)
	require.NoError(t, err)

	s, err := sessionStorage.GetSession(session.Id)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, session, s)

	sessionIds, err := sessionStorage.GetSessionsForModel(session.ModelID)
	require.NoError(t, err)
	require.Equal(t, []string{session.Id}, sessionIds)
}

func TestRemoveSession(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	session := &Session{
		Id:               "0x0",
		UserAddr:         "0x1",
		ProviderAddr:     "0x2",
		EndsAt:           big.NewInt(100),
		ModelID:          "0x3",
		TPSScaled1000Arr: []int{1, 2, 3},
		TTFTMsArr:        []int{4, 5, 6},
		FailoverEnabled:  true,
	}

	err := sessionStorage.AddSession(session)
	require.NoError(t, err)

	err = sessionStorage.RemoveSession(session.Id)
	require.NoError(t, err)

	s, err := sessionStorage.GetSession(session.Id)
	require.NoError(t, err)
	require.Nil(t, s)

	sessionIds, err := sessionStorage.GetSessionsForModel(session.ModelID)
	require.NoError(t, err)
	require.Empty(t, sessionIds)
}

func TestActivityStorage_AddAndGetActivities(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	modelID := "0xabc"
	a1 := &PromptActivity{SessionID: "0x1", StartTime: 100, EndTime: 101}
	a2 := &PromptActivity{SessionID: "0x2", StartTime: 200, EndTime: 201}

	require.NoError(t, sessionStorage.AddActivity(modelID, a1))
	require.NoError(t, sessionStorage.AddActivity(modelID, a2))

	activities, err := sessionStorage.GetActivities(modelID)
	require.NoError(t, err)
	require.Len(t, activities, 2)

	got := map[string]PromptActivity{}
	for _, a := range activities {
		got[a.SessionID] = *a
	}
	require.Equal(t, *a1, got[a1.SessionID])
	require.Equal(t, *a2, got[a2.SessionID])
}

func TestActivityStorage_RemoveOldActivities(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	modelID := "0xdef"
	oldActivity := &PromptActivity{SessionID: "0xold", StartTime: 100, EndTime: 110}
	newActivity := &PromptActivity{SessionID: "0xnew", StartTime: 200, EndTime: 210}

	require.NoError(t, sessionStorage.AddActivity(modelID, oldActivity))
	require.NoError(t, sessionStorage.AddActivity(modelID, newActivity))

	require.NoError(t, sessionStorage.RemoveOldActivities(modelID, 150))

	activities, err := sessionStorage.GetActivities(modelID)
	require.NoError(t, err)
	require.Len(t, activities, 1)
	require.Equal(t, newActivity.SessionID, activities[0].SessionID)
	require.Equal(t, newActivity.EndTime, activities[0].EndTime)
}

func TestActivityStorage_IgnoresLegacyArrayKey(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	modelID := "0xlegacy"
	legacyValue, err := json.Marshal([]*PromptActivity{
		{SessionID: "0xlegacy", StartTime: 1, EndTime: 2},
	})
	require.NoError(t, err)

	legacyKey := []byte("activity:" + modelID)
	require.NoError(t, storage.SetWithTTL(legacyKey, legacyValue, time.Hour))

	activities, err := sessionStorage.GetActivities(modelID)
	require.NoError(t, err)
	require.Empty(t, activities)
}

func TestAddUserBoundEnforcesKeyToAddress(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	pub := lib.HexString(crypto.FromECDSAPub(key.Public().(*ecdsa.PublicKey))).Hex()
	addr := crypto.PubkeyToAddress(key.PublicKey)

	// matching key is stored
	require.NoError(t, sessionStorage.AddUserBound(&User{Addr: addr.Hex(), PubKey: pub}))
	got, err := sessionStorage.GetUser(addr.Hex())
	require.NoError(t, err)
	require.Equal(t, pub, got.PubKey)

	// a key that derives to a different address is refused and does not
	// disturb the stored record
	other, err := crypto.GenerateKey()
	require.NoError(t, err)
	otherPub := lib.HexString(crypto.FromECDSAPub(other.Public().(*ecdsa.PublicKey))).Hex()
	require.Error(t, sessionStorage.AddUserBound(&User{Addr: addr.Hex(), PubKey: otherPub}))
	got, err = sessionStorage.GetUser(addr.Hex())
	require.NoError(t, err)
	require.Equal(t, pub, got.PubKey)

	// garbage key is refused
	require.Error(t, sessionStorage.AddUserBound(&User{Addr: addr.Hex(), PubKey: "0xdeadbeef"}))

	// the unbound path still permits the consumer-side contract-provider
	// shape (owner key under a contract address)
	require.NoError(t, sessionStorage.AddUser(&User{Addr: "0x000000000000000000000000000000000000c0de", PubKey: otherPub}))
}

func TestAddUserOverwriteReplacesStaleRecord(t *testing.T) {
	storage := NewTestStorage()
	sessionStorage := NewSessionStorage(storage)

	u1 := &User{Addr: "0xAbC", PubKey: "0x1111"}
	require.NoError(t, sessionStorage.AddUser(u1))

	// same key is idempotent
	require.NoError(t, sessionStorage.AddUser(&User{Addr: "0xabc", PubKey: "0x1111"}))

	// ownership is proven by the caller before AddUser; a later write wins so
	// a wallet owner can replace a stale record
	require.NoError(t, sessionStorage.AddUser(&User{Addr: "0xabc", PubKey: "0x2222"}))

	got, err := sessionStorage.GetUser("0xabc")
	require.NoError(t, err)
	require.Equal(t, "0x2222", got.PubKey)
}
