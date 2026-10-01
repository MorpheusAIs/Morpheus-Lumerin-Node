package proxyapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSessionUsesCreateTempUnderStorageDir(t *testing.T) {
	sm := NewStreamingSessionManager()
	session, err := sm.CreateSession("0xabc", 2, 100, "audio/mpeg")
	require.NoError(t, err)
	require.NotEmpty(t, session.StreamID)
	require.True(t, IsManagedAudioPath(session.TempFilePath))
	require.True(t, strings.HasPrefix(filepath.Base(session.TempFilePath), "audiostream-"))
	// client-looking id must not appear in path
	require.NotContains(t, session.TempFilePath, "0xabc")
	require.FileExists(t, session.TempFilePath)
	sm.RemoveSession(session.StreamID)
	_, err = os.Stat(session.TempFilePath)
	require.True(t, os.IsNotExist(err))
}

func TestCreateSessionGeneratesUniqueStreamIDs(t *testing.T) {
	sm := NewStreamingSessionManager()
	a, err := sm.CreateSession("0x1", 1, 1, "audio/wav")
	require.NoError(t, err)
	b, err := sm.CreateSession("0x1", 1, 1, "audio/wav")
	require.NoError(t, err)
	require.NotEqual(t, a.StreamID, b.StreamID)
	sm.RemoveSession(a.StreamID)
	sm.RemoveSession(b.StreamID)
}

func TestSafeRemoveManagedAudioPathRefusesClientPath(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "client-supplied.txt")
	require.NoError(t, os.WriteFile(outside, []byte("x"), 0o600))
	err := SafeRemoveManagedAudioPath(outside)
	require.Error(t, err)
	require.FileExists(t, outside)
}

func TestSessionOwnsStream(t *testing.T) {
	sm := NewStreamingSessionManager()
	s, err := sm.CreateSession("0xSESSION", 1, 1, "audio/mpeg")
	require.NoError(t, err)
	require.True(t, sm.SessionOwnsStream(s.StreamID, "0xSESSION"))
	require.True(t, sm.SessionOwnsStream(s.StreamID, "0xsession"))
	require.False(t, sm.SessionOwnsStream(s.StreamID, "0xOTHER"))
	sm.RemoveSession(s.StreamID)
}
