package proxyapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	AUDIO_STREAM_TIMEOUT_SECONDS = 1 * 60 * 60 // 1 hour timeout for audio streaming sessions
	audioStreamDirName           = "morpheus-audio-streams"
)

// AudioStreamStorageDir is the fixed directory for server-created audio stream temp files.
func AudioStreamStorageDir() string {
	return filepath.Join(os.TempDir(), audioStreamDirName)
}

// IsManagedAudioPath reports whether path is under the audio stream storage dir.
func IsManagedAudioPath(path string) bool {
	return lib.PathUnderDir(path, AudioStreamStorageDir())
}

// SafeRemoveManagedAudioPath removes path only if it is under the managed audio dir.
func SafeRemoveManagedAudioPath(path string) error {
	if !IsManagedAudioPath(path) {
		return fmt.Errorf("refuse remove on non-managed path")
	}
	return os.Remove(path)
}

// StreamingSession represents an active audio streaming session
type StreamingSession struct {
	StreamID     string
	SessionID    string
	TotalChunks  uint32
	FileSize     uint64
	ContentType  string
	TempFilePath string // Path to temporary file where chunks are written
	ChunkCount   uint32
	StartTime    time.Time
	LastActivity time.Time
}

// StreamingSessionManager manages active streaming sessions
type StreamingSessionManager struct {
	sessions   map[string]*StreamingSession
	storageDir string
	mu         sync.Mutex
}

func NewStreamingSessionManager() *StreamingSessionManager {
	return &StreamingSessionManager{
		sessions:   make(map[string]*StreamingSession),
		storageDir: AudioStreamStorageDir(),
	}
}

func generateStreamID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession allocates a server-generated streamID and a CreateTemp file under storageDir.
// Client-supplied stream IDs are not used for paths or map keys.
func (sm *StreamingSessionManager) CreateSession(sessionID string, totalChunks uint32, fileSize uint64, contentType string) (*StreamingSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	streamID, err := generateStreamID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate stream id: %w", err)
	}

	tempFilePath, err := sm.createTempFile(contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	session := &StreamingSession{
		StreamID:     streamID,
		SessionID:    sessionID,
		TotalChunks:  totalChunks,
		FileSize:     fileSize,
		ContentType:  contentType,
		TempFilePath: tempFilePath,
		ChunkCount:   0,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
	}
	sm.sessions[streamID] = session
	return session, nil
}

func (sm *StreamingSessionManager) createTempFile(contentType string) (string, error) {
	if err := os.MkdirAll(sm.storageDir, 0o700); err != nil {
		return "", err
	}

	extension := getFileExtensionFromContentType(contentType)
	pattern := "audiostream-*" + extension
	f, err := os.CreateTemp(sm.storageDir, pattern)
	if err != nil {
		return "", err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", err
	}

	if !lib.PathUnderDir(name, sm.storageDir) {
		_ = os.Remove(name)
		return "", fmt.Errorf("temp path not under storage dir")
	}
	return name, nil
}

func (sm *StreamingSessionManager) GetSession(streamID string) (*StreamingSession, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session, exists := sm.sessions[streamID]
	return session, exists
}

func (sm *StreamingSessionManager) RemoveSession(streamID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[streamID]; exists {
		if session.TempFilePath != "" {
			_ = SafeRemoveManagedAudioPath(session.TempFilePath)
		}
		delete(sm.sessions, streamID)
	}
}

func (sm *StreamingSessionManager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	now := time.Now()
	for streamID, session := range sm.sessions {
		if now.Sub(session.LastActivity).Seconds() > AUDIO_STREAM_TIMEOUT_SECONDS {
			if session.TempFilePath != "" {
				_ = SafeRemoveManagedAudioPath(session.TempFilePath)
			}
			delete(sm.sessions, streamID)
		}
	}
}

// SessionOwnsStream reports whether streamID is bound to sessionID (case-insensitive hex).
func (sm *StreamingSessionManager) SessionOwnsStream(streamID, sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session, exists := sm.sessions[streamID]
	if !exists {
		return false
	}
	return strings.EqualFold(session.SessionID, sessionID)
}

func getFileExtensionFromContentType(contentType string) string {
	extensions := map[string]string{
		"audio/mpeg":     ".mp3",
		"audio/mp3":      ".mp3",
		"audio/wav":      ".wav",
		"audio/wave":     ".wav",
		"audio/x-wav":    ".wav",
		"audio/vnd.wave": ".wav",
		"audio/ogg":      ".ogg",
		"audio/flac":     ".flac",
		"audio/aac":      ".aac",
		"audio/mp4":      ".m4a",
		"audio/x-m4a":    ".m4a",
		"audio/webm":     ".webm",
		"audio/opus":     ".opus",
		"audio/x-ms-wma": ".wma",
		"audio/amr":      ".amr",
		"audio/3gpp":     ".3gp",
		"audio/x-aiff":   ".aiff",
		"audio/aiff":     ".aiff",
	}

	if ext, exists := extensions[contentType]; exists {
		return ext
	}
	return ".mp3" // default extension
}
