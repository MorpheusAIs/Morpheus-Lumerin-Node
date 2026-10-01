package proxyapi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	AUDIO_STREAM_TIMEOUT_SECONDS = 1 * 60 * 60 // 1 hour timeout for audio streaming sessions
	audioStreamDirName           = "morpheus-audio-streams"

	// Bounds on a single audio stream. Consumers send 1 MiB chunks
	// (SENDER_AUDIO_STREAM_CHUNK_SIZE); these caps are generous relative to that
	// while keeping the on-disk footprint of any one stream finite.
	MAX_AUDIO_STREAM_FILE_SIZE    = 256 * 1024 * 1024 // 256 MiB total per stream
	MAX_AUDIO_STREAM_CHUNK_SIZE   = 8 * 1024 * 1024   // 8 MiB per decoded chunk
	MAX_AUDIO_STREAM_TOTAL_CHUNKS = 4096
	MAX_AUDIO_STREAMS_PER_SESSION = 4
)

// Client-supplied stream identifiers are used only as an in-memory lookup key,
// never for filesystem paths. Restrict them to a conservative charset and length.
var streamIDRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

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

// ValidateStreamID checks that a client-supplied stream identifier is safe to
// use as a lookup key.
func ValidateStreamID(streamID string) error {
	if !streamIDRegexp.MatchString(streamID) {
		return fmt.Errorf("invalid stream id")
	}
	return nil
}

// StreamingSession represents an active audio streaming session
type StreamingSession struct {
	StreamID     string // client-supplied identifier, echoed in responses; never used in paths
	SessionID    string
	TotalChunks  uint32
	FileSize     uint64
	ContentType  string
	TempFilePath string // Path to temporary file where chunks are written (server-generated)
	ChunkCount   uint32
	BytesWritten uint64
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

// streamKey namespaces the client stream identifier by the owning on-chain
// session so one session can neither collide with nor address another's stream.
func streamKey(sessionID, streamID string) string {
	return strings.ToLower(sessionID) + ":" + streamID
}

// CreateSession registers a stream under (sessionID, streamID) and allocates a
// server-generated temp file under storageDir. The client-supplied streamID is
// kept for wire compatibility (consumers reuse it on chunk/end) but is only an
// in-memory key; the filesystem path never derives from it.
func (sm *StreamingSessionManager) CreateSession(streamID, sessionID string, totalChunks uint32, fileSize uint64, contentType string) (*StreamingSession, error) {
	if err := ValidateStreamID(streamID); err != nil {
		return nil, err
	}
	if totalChunks == 0 || totalChunks > MAX_AUDIO_STREAM_TOTAL_CHUNKS {
		return nil, fmt.Errorf("total chunks out of range")
	}
	if fileSize == 0 || fileSize > MAX_AUDIO_STREAM_FILE_SIZE {
		return nil, fmt.Errorf("file size out of range")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := streamKey(sessionID, streamID)
	if _, exists := sm.sessions[key]; exists {
		return nil, fmt.Errorf("stream with ID %s already exists", streamID)
	}

	active := 0
	for _, s := range sm.sessions {
		if strings.EqualFold(s.SessionID, sessionID) {
			active++
		}
	}
	if active >= MAX_AUDIO_STREAMS_PER_SESSION {
		return nil, fmt.Errorf("too many concurrent streams for session")
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
	sm.sessions[key] = session
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

// GetSession returns the stream registered under (sessionID, streamID). A
// stream is only visible to the on-chain session that created it.
func (sm *StreamingSessionManager) GetSession(streamID, sessionID string) (*StreamingSession, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session, exists := sm.sessions[streamKey(sessionID, streamID)]
	return session, exists
}

// AppendChunk validates and appends decoded chunk data to the stream's temp
// file, enforcing ordering and size bounds.
func (sm *StreamingSessionManager) AppendChunk(streamID, sessionID string, chunkIndex uint32, data []byte) (*StreamingSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[streamKey(sessionID, streamID)]
	if !exists {
		return nil, fmt.Errorf("streaming session %s not found", streamID)
	}
	if chunkIndex != session.ChunkCount {
		return nil, fmt.Errorf("expected chunk index %d, got %d", session.ChunkCount, chunkIndex)
	}
	if session.ChunkCount >= session.TotalChunks {
		return nil, fmt.Errorf("stream already has all %d chunks", session.TotalChunks)
	}
	if uint64(len(data)) > MAX_AUDIO_STREAM_CHUNK_SIZE {
		return nil, fmt.Errorf("chunk too large")
	}
	if session.BytesWritten+uint64(len(data)) > session.FileSize {
		return nil, fmt.Errorf("stream exceeds declared file size")
	}
	if !lib.PathUnderDir(session.TempFilePath, sm.storageDir) {
		return nil, fmt.Errorf("temp path not under storage dir")
	}

	file, err := os.OpenFile(session.TempFilePath, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file for writing: %w", err)
	}
	defer file.Close()

	n, err := file.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write chunk data to temp file: %w", err)
	}

	session.BytesWritten += uint64(n)
	session.ChunkCount++
	session.LastActivity = time.Now()
	return session, nil
}

func (sm *StreamingSessionManager) RemoveSession(streamID, sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	key := streamKey(sessionID, streamID)
	if session, exists := sm.sessions[key]; exists {
		if session.TempFilePath != "" {
			_ = SafeRemoveManagedAudioPath(session.TempFilePath)
		}
		delete(sm.sessions, key)
	}
}

func (sm *StreamingSessionManager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	now := time.Now()
	for key, session := range sm.sessions {
		if now.Sub(session.LastActivity).Seconds() > AUDIO_STREAM_TIMEOUT_SECONDS {
			if session.TempFilePath != "" {
				_ = SafeRemoveManagedAudioPath(session.TempFilePath)
			}
			delete(sm.sessions, key)
		}
	}
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
