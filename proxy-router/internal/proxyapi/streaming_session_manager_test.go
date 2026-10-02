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
	session, err := sm.CreateSession("abc123", "0xSESSION", 2, 100, "audio/mpeg")
	require.NoError(t, err)
	require.Equal(t, "abc123", session.StreamID)
	require.True(t, IsManagedAudioPath(session.TempFilePath))
	require.True(t, strings.HasPrefix(filepath.Base(session.TempFilePath), "audiostream-"))
	require.True(t, strings.HasSuffix(session.TempFilePath, ".mp3"))
	// client id must not appear in the path
	require.NotContains(t, filepath.Base(session.TempFilePath), "abc123")
	require.FileExists(t, session.TempFilePath)
	sm.RemoveSession(session.StreamID, "0xSESSION")
	_, err = os.Stat(session.TempFilePath)
	require.True(t, os.IsNotExist(err))
}

func TestCreateSessionRejectsUnsafeStreamIDs(t *testing.T) {
	sm := NewStreamingSessionManager()
	dd := string([]byte{46, 46})
	bad := []string{
		"",
		dd + "/x",
		"a/b",
		"a\\b",
		"a b",
		"a.b",
		strings.Repeat("a", 65),
	}
	for _, id := range bad {
		_, err := sm.CreateSession(id, "0xS", 1, 1, "audio/wav")
		require.Error(t, err, "id %q should be rejected", id)
	}
	// what the consumer actually sends: 32 lowercase hex chars
	ok, err := sm.CreateSession("0123456789abcdef0123456789abcdef", "0xS", 1, 1, "audio/wav")
	require.NoError(t, err)
	sm.RemoveSession(ok.StreamID, "0xS")
}

func TestCreateSessionRejectsDuplicateWithinSession(t *testing.T) {
	sm := NewStreamingSessionManager()
	a, err := sm.CreateSession("same", "0xS", 1, 1, "audio/wav")
	require.NoError(t, err)
	_, err = sm.CreateSession("same", "0xS", 1, 1, "audio/wav")
	require.Error(t, err)
	// same client id under a different session is a different stream
	b, err := sm.CreateSession("same", "0xOTHER", 1, 1, "audio/wav")
	require.NoError(t, err)
	require.NotEqual(t, a.TempFilePath, b.TempFilePath)
	sm.RemoveSession("same", "0xS")
	sm.RemoveSession("same", "0xOTHER")
}

func TestGetSessionIsScopedToOwningSession(t *testing.T) {
	sm := NewStreamingSessionManager()
	s, err := sm.CreateSession("abc", "0xSESSION", 1, 1, "audio/mpeg")
	require.NoError(t, err)
	_, ok := sm.GetSession(s.StreamID, "0xSESSION")
	require.True(t, ok)
	_, ok = sm.GetSession(s.StreamID, "0xsession") // hex case-insensitive
	require.True(t, ok)
	_, ok = sm.GetSession(s.StreamID, "0xOTHER")
	require.False(t, ok)
	// another session cannot remove it either
	sm.RemoveSession(s.StreamID, "0xOTHER")
	require.FileExists(t, s.TempFilePath)
	sm.RemoveSession(s.StreamID, "0xSESSION")
}

func TestCreateSessionEnforcesBounds(t *testing.T) {
	sm := NewStreamingSessionManager()
	_, err := sm.CreateSession("a", "0xS", 0, 1, "audio/wav")
	require.Error(t, err)
	_, err = sm.CreateSession("a", "0xS", MAX_AUDIO_STREAM_TOTAL_CHUNKS+1, 1, "audio/wav")
	require.Error(t, err)
	_, err = sm.CreateSession("a", "0xS", 1, 0, "audio/wav")
	require.Error(t, err)
	_, err = sm.CreateSession("a", "0xS", 1, MAX_AUDIO_STREAM_FILE_SIZE+1, "audio/wav")
	require.Error(t, err)

	// more chunks than bytes is impossible
	_, err = sm.CreateSession("a", "0xS", 11, 10, "audio/wav")
	require.Error(t, err)

	for i := 0; i < MAX_AUDIO_STREAMS_PER_SESSION; i++ {
		_, err := sm.CreateSession("s"+string(rune('a'+i)), "0xS", 1, 1, "audio/wav")
		require.NoError(t, err)
	}
	_, err = sm.CreateSession("overflow", "0xS", 1, 1, "audio/wav")
	require.Error(t, err)
	for i := 0; i < MAX_AUDIO_STREAMS_PER_SESSION; i++ {
		sm.RemoveSession("s"+string(rune('a'+i)), "0xS")
	}
}

func TestCreateSessionEnforcesGlobalInflightCap(t *testing.T) {
	sm := NewStreamingSessionManager()
	// declared sizes only; CreateTemp files are empty so this costs no disk
	perStream := uint64(MAX_AUDIO_STREAM_FILE_SIZE)
	n := int(uint64(MAX_AUDIO_STREAM_INFLIGHT_BYTES) / perStream) // 8
	for i := 0; i < n; i++ {
		sid := "0xS" + string(rune('a'+i)) // distinct sessions avoid the per-session cap
		_, err := sm.CreateSession("x", sid, 1, perStream, "audio/wav")
		require.NoError(t, err, "stream %d", i)
	}
	_, err := sm.CreateSession("x", "0xOVER", 1, 1, "audio/wav")
	require.Error(t, err)
	sm.RemoveSession("x", "0xSa")
	_, err = sm.CreateSession("x", "0xOVER", 1, 1, "audio/wav")
	require.NoError(t, err)
	for i := 1; i < n; i++ {
		sm.RemoveSession("x", "0xS"+string(rune('a'+i)))
	}
	sm.RemoveSession("x", "0xOVER")
}

func TestAppendChunkEnforcesOrderAndSize(t *testing.T) {
	sm := NewStreamingSessionManager()
	s, err := sm.CreateSession("abc", "0xS", 2, 10, "audio/wav")
	require.NoError(t, err)
	defer sm.RemoveSession("abc", "0xS")

	// wrong session
	_, err = sm.AppendChunk("abc", "0xOTHER", 0, []byte("12345"))
	require.Error(t, err)

	// out of order
	_, err = sm.AppendChunk("abc", "0xS", 1, []byte("12345"))
	require.Error(t, err)

	// ok
	s, err = sm.AppendChunk("abc", "0xS", 0, []byte("12345"))
	require.NoError(t, err)
	require.Equal(t, uint32(1), s.ChunkCount)
	require.Equal(t, uint64(5), s.BytesWritten)

	// exceeds declared size
	_, err = sm.AppendChunk("abc", "0xS", 1, []byte("123456"))
	require.Error(t, err)

	// fits exactly
	s, err = sm.AppendChunk("abc", "0xS", 1, []byte("12345"))
	require.NoError(t, err)
	require.Equal(t, uint32(2), s.ChunkCount)

	// no more chunks accepted
	_, err = sm.AppendChunk("abc", "0xS", 2, []byte("x"))
	require.Error(t, err)

	data, err := os.ReadFile(s.TempFilePath)
	require.NoError(t, err)
	require.Equal(t, "1234512345", string(data))
}

func TestSafeRemoveManagedAudioPathRefusesClientPath(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "client-supplied.txt")
	require.NoError(t, os.WriteFile(outside, []byte("x"), 0o600))
	err := SafeRemoveManagedAudioPath(outside)
	require.Error(t, err)
	require.FileExists(t, outside)
}

func TestSafeAudioExtension(t *testing.T) {
	require.Equal(t, ".mp3", safeAudioExtension("Voice Memo.MP3"))
	require.Equal(t, ".wav", safeAudioExtension("/some/where/clip.wav"))
	require.Equal(t, "", safeAudioExtension("evil.sh"))
	require.Equal(t, "", safeAudioExtension("noext"))
	require.Equal(t, "", safeAudioExtension(""))
}
