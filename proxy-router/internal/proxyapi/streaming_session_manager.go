package proxyapi

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	AUDIO_STREAM_TIMEOUT_SECONDS = 1 * 60 * 60 // 1 hour timeout for audio streaming sessions
)

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
	sessions map[string]*StreamingSession
	mu       sync.RWMutex // Protects concurrent access to sessions map
}

func NewStreamingSessionManager() *StreamingSessionManager {
	return &StreamingSessionManager{
		sessions: make(map[string]*StreamingSession),
	}
}

func (sm *StreamingSessionManager) CreateSession(streamID, sessionID string, totalChunks uint32, fileSize uint64, contentType string) (*StreamingSession, error) {
	// Create temporary file for streaming chunks
	tempFilePath, err := sm.createTempFile(streamID, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %s", err)
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

	sm.mu.Lock()
	sm.sessions[streamID] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *StreamingSessionManager) createTempFile(streamID, contentType string) (string, error) {
	// Create temporary file
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%d_stream_%s", time.Now().UnixNano(), streamID))

	// Detect file extension from content type
	extension := getFileExtensionFromContentType(contentType)
	if extension != "" {
		tempFilePath += extension
	}

	// Create the file
	file, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	file.Close() // Close immediately, we'll open it for appending when needed

	return tempFilePath, nil
}

func (sm *StreamingSessionManager) GetSession(streamID string) (*StreamingSession, bool) {
	sm.mu.RLock()
	session, exists := sm.sessions[streamID]
	sm.mu.RUnlock()
	return session, exists
}

func (sm *StreamingSessionManager) RemoveSession(streamID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[streamID]; exists {
		// Clean up temporary file
		if session.TempFilePath != "" {
			os.Remove(session.TempFilePath)
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
			// Clean up temporary file
			if session.TempFilePath != "" {
				os.Remove(session.TempFilePath)
			}
			delete(sm.sessions, streamID)
		}
	}
}

// getFileExtensionFromContentType returns the appropriate file extension for a given content type
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
