package storages

import (
	"math/big"
	"testing"

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

	s, ok := sessionStorage.GetSession(session.Id)
	require.True(t, ok)
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

	_, ok := sessionStorage.GetSession(session.Id)
	require.False(t, ok)

	sessionIds, err := sessionStorage.GetSessionsForModel(session.ModelID)
	require.NoError(t, err)
	require.Empty(t, sessionIds)
}
