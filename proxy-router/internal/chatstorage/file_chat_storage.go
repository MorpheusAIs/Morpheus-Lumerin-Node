package chatstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
)

// ChatStorage handles storing conversations to files.
type ChatStorage struct {
	dirPath            string                 // Directory path to store the files
	fileMutexes        map[string]*sync.Mutex // Map to store mutexes for each file
	fileMutexesMu      sync.Mutex             // Protects fileMutexes
	forwardChatContext bool
}

const (
	chatDirectoryMode os.FileMode = 0o700
	chatFileMode      os.FileMode = 0o600
)

// NewChatStorage creates a new instance of ChatStorage.
func NewChatStorage(dirPath string) *ChatStorage {
	return &ChatStorage{
		dirPath:     dirPath,
		fileMutexes: make(map[string]*sync.Mutex),
	}
}

// StorePromptResponseToFile stores the prompt and response to a file.
func (cs *ChatStorage) StorePromptResponseToFile(identifier string, isLocal bool, modelId string, prompt interface{}, responses []gcs.Chunk, promptAt time.Time, responseAt time.Time) error {
	if err := ensurePrivateDirectory(cs.dirPath); err != nil {
		return err
	}

	filePath := filepath.Join(cs.dirPath, identifier+".json")
	fileMutex := cs.getFileMutex(filePath)

	// Lock the file mutex
	fileMutex.Lock()
	defer fileMutex.Unlock()

	var chatHistory gcs.ChatHistory
	fileContent, err := readPrivateFile(filePath)
	if err == nil {
		if err := json.Unmarshal(fileContent, &chatHistory); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	resps := make([]string, len(responses))
	for i, r := range responses {
		resps[i] = r.String()
	}

	isImageContent := false
	isVideoRawContent := false
	isAudioContent := false
	if len(responses) > 0 {
		isImageContent = responses[0].Type() == gcs.ChunkTypeImage
		isVideoRawContent = responses[0].Type() == gcs.ChunkTypeVideo
		isAudioContent = responses[0].Type() == gcs.ChunkTypeAudioTranscriptionText ||
			responses[0].Type() == gcs.ChunkTypeAudioTranscriptionJson ||
			responses[0].Type() == gcs.ChunkTypeAudioTranscriptionDelta
	}

	var newEntry gcs.ChatMessage
	var title string

	switch p := prompt.(type) {
	case *gcs.OpenAICompletionRequestExtra:
		newEntry = gcs.ChatMessage{
			Prompt:            prompt,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
		title = p.Messages[0].Content
	case *gcs.AudioTranscriptionRequest:
		// Store audio transcription request directly
		newEntry = gcs.ChatMessage{
			Prompt:            p,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
		// Use a default title for audio transcription or the prompt if available
		if p.Prompt != "" {
			title = "Audio Transcription: " + p.Prompt
		} else {
			title = "Audio Transcription"
		}
	case *gcs.EmbeddingsRequest:
		newEntry = gcs.ChatMessage{
			Prompt:            p,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
	case *gcs.AudioSpeechRequest:
		// Store audio speech request directly
		newEntry = gcs.ChatMessage{
			Prompt:            p,
			Response:          strings.Join(resps, ""),
			PromptAt:          promptAt.Unix(),
			ResponseAt:        responseAt.Unix(),
			IsImageContent:    isImageContent,
			IsVideoRawContent: isVideoRawContent,
			IsAudioContent:    isAudioContent,
		}
	default:
		return fmt.Errorf("unsupported prompt type: %T", prompt)
	}

	if chatHistory.Messages == nil && len(chatHistory.Messages) == 0 {
		chatHistory.ModelId = modelId
		chatHistory.Title = title
		chatHistory.IsLocal = isLocal
	}

	newMessages := append(chatHistory.Messages, newEntry)
	chatHistory.Messages = newMessages

	updatedContent, err := json.MarshalIndent(chatHistory, "", "  ")
	if err != nil {
		return err
	}

	if err := writePrivateFileAtomically(filePath, updatedContent); err != nil {
		return err
	}

	return nil
}

func (cs *ChatStorage) GetChats() []gcs.Chat {
	var chats []gcs.Chat
	if err := ensurePrivateDirectory(cs.dirPath); err != nil {
		return chats
	}
	files, err := os.ReadDir(cs.dirPath)
	if err != nil {
		return chats
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		chatID := strings.TrimSuffix(file.Name(), ".json")

		fileContent, err := cs.LoadChatFromFile(chatID)
		if err != nil {
			continue
		}
		if len(fileContent.Messages) == 0 {
			continue
		}
		chats = append(chats, gcs.Chat{
			ChatID:    chatID,
			Title:     fileContent.Title,
			CreatedAt: fileContent.Messages[0].PromptAt,
			ModelID:   fileContent.ModelId,
			IsLocal:   fileContent.IsLocal,
		})
	}

	return chats
}

func (cs *ChatStorage) DeleteChat(identifier string) error {
	if err := ensurePrivateDirectory(cs.dirPath); err != nil {
		return err
	}
	filePath := filepath.Join(cs.dirPath, identifier+".json")
	fileMutex := cs.getFileMutex(filePath)

	fileMutex.Lock()
	defer fileMutex.Unlock()

	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}

func (cs *ChatStorage) UpdateChatTitle(identifier string, title string) error {
	if err := ensurePrivateDirectory(cs.dirPath); err != nil {
		return err
	}

	filePath := filepath.Join(cs.dirPath, identifier+".json")
	fileMutex := cs.getFileMutex(filePath)

	fileMutex.Lock()
	defer fileMutex.Unlock()

	fileContent, err := readPrivateFile(filePath)
	if err != nil {
		return err
	}

	var chat gcs.ChatHistory
	if err := json.Unmarshal(fileContent, &chat); err != nil {
		return err
	}
	chat.Title = title

	updatedContent, err := json.MarshalIndent(&chat, "", "  ")
	if err != nil {
		return err
	}

	if err := writePrivateFileAtomically(filePath, updatedContent); err != nil {
		return err
	}

	return nil
}

func (cs *ChatStorage) LoadChatFromFile(identifier string) (*gcs.ChatHistory, error) {
	if err := ensurePrivateDirectory(cs.dirPath); err != nil {
		return nil, err
	}
	filePath := filepath.Join(cs.dirPath, identifier+".json")
	fileMutex := cs.getFileMutex(filePath)

	fileMutex.Lock()
	defer fileMutex.Unlock()

	var data gcs.ChatHistory
	fileContent, err := readPrivateFile(filePath)
	if err != nil {
		return &data, err
	}

	if err := json.Unmarshal(fileContent, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// getFileMutex initializes and returns the mutex for a file. The map itself is
// protected because requests for different chats can arrive concurrently.
func (cs *ChatStorage) getFileMutex(filePath string) *sync.Mutex {
	cs.fileMutexesMu.Lock()
	defer cs.fileMutexesMu.Unlock()

	if fileMutex, exists := cs.fileMutexes[filePath]; exists {
		return fileMutex
	}

	fileMutex := &sync.Mutex{}
	cs.fileMutexes[filePath] = fileMutex
	return fileMutex
}

func ensurePrivateDirectory(dirPath string) error {
	if err := os.MkdirAll(dirPath, chatDirectoryMode); err != nil {
		return err
	}
	return os.Chmod(dirPath, chatDirectoryMode)
}

func readPrivateFile(filePath string) ([]byte, error) {
	if err := os.Chmod(filePath, chatFileMode); err != nil {
		return nil, err
	}
	return os.ReadFile(filePath)
}

// writePrivateFileAtomically writes a private temporary file beside its final
// destination, flushes and closes it, then replaces the destination with a
// same-directory rename. Readers therefore never observe partially-written
// JSON.
func writePrivateFileAtomically(filePath string, content []byte) error {
	if err := ensurePrivateDirectory(filepath.Dir(filePath)); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "."+filepath.Base(filePath)+".tmp-")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if err := tempFile.Chmod(chatFileMode); err != nil {
		return err
	}
	if _, err := tempFile.Write(content); err != nil {
		return err
	}
	if err := tempFile.Sync(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		return err
	}

	return os.Chmod(filePath, chatFileMode)
}
