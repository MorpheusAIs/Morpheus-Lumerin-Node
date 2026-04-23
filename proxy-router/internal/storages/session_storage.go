package storages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// Default TTL values for stored data
const (
	ActivityTTL = 24 * time.Hour
	SessionTTL  = 30 * 24 * time.Hour // 30 days; sessions are also cleaned by expiry handler
)

type SessionStorage struct {
	db *Storage
}

func NewSessionStorage(storage *Storage) *SessionStorage {
	return &SessionStorage{
		db: storage,
	}
}

// GetUser retrieves a user by address. Returns (nil, nil) if not found.
// Returns a non-nil error only on actual storage or deserialization failures.
func (s *SessionStorage) GetUser(addr string) (*User, error) {
	addr = strings.ToLower(addr)
	key := fmt.Sprintf("user:%s", addr)
	userJson, err := s.db.Get([]byte(key))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading user %s: %w", addr, err)
	}

	user := &User{}
	if err := json.Unmarshal(userJson, user); err != nil {
		return nil, fmt.Errorf("error unmarshaling user %s: %w", addr, err)
	}

	return user, nil
}

func (s *SessionStorage) AddUser(user *User) error {
	addr := strings.ToLower(user.Addr)
	key := fmt.Sprintf("user:%s", addr)
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.db.Set([]byte(key), userJson)
}

// GetSession retrieves a session by ID. Returns (nil, nil) if not found.
// Returns a non-nil error only on actual storage or deserialization failures.
func (s *SessionStorage) GetSession(id string) (*Session, error) {
	sessionJson, err := s.db.Get(formatSessionKey(id))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading session %s: %w", id, err)
	}

	session := &Session{}
	if err := json.Unmarshal(sessionJson, session); err != nil {
		return nil, fmt.Errorf("error unmarshaling session %s: %w", id, err)
	}

	return session, nil
}

// AddSession atomically stores a session and its model-session index in a single transaction.
func (s *SessionStorage) AddSession(session *Session) error {
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.db.RunInTransaction(func(txn *badger.Txn) error {
		sessionEntry := badger.NewEntry(formatSessionKey(session.Id), sessionJson).WithTTL(SessionTTL)
		if err := txn.SetEntry(sessionEntry); err != nil {
			return fmt.Errorf("error storing session %s: %w", session.Id, err)
		}

		indexEntry := badger.NewEntry(formatModelSessionKey(session.ModelID, session.Id), []byte{}).WithTTL(SessionTTL)
		if err := txn.SetEntry(indexEntry); err != nil {
			return fmt.Errorf("error storing model-session index for %s: %w", session.Id, err)
		}

		return nil
	})
}

// RemoveSession atomically removes a session and its model-session index.
func (s *SessionStorage) RemoveSession(id string) error {
	ses, err := s.GetSession(id)
	if err != nil {
		return fmt.Errorf("error looking up session %s for removal: %w", id, err)
	}
	if ses == nil {
		return nil
	}

	sessionId := strings.ToLower(id)
	return s.db.RunInTransaction(func(txn *badger.Txn) error {
		sessionKey := fmt.Sprintf("session:%s", sessionId)
		if err := txn.Delete([]byte(sessionKey)); err != nil {
			return fmt.Errorf("error deleting session %s: %w", sessionId, err)
		}

		modelSessionKey := formatModelSessionKey(ses.ModelID, sessionId)
		if err := txn.Delete(modelSessionKey); err != nil {
			return fmt.Errorf("error deleting model-session index for %s: %w", sessionId, err)
		}

		return nil
	})
}

// GetSessions returns all sessions using a single transaction to fetch keys and values.
// Accepts an optional context for cancellation during large scans.
func (s *SessionStorage) GetSessions(ctxOpts ...context.Context) ([]Session, error) {
	prefix := formatSessionKey("")
	_, values, err := s.db.GetPrefixWithValues(prefix, ctxOpts...)
	if err != nil {
		return nil, fmt.Errorf("error reading sessions: %w", err)
	}

	sessions := make([]Session, 0, len(values))
	for _, val := range values {
		session := Session{}
		if err := json.Unmarshal(val, &session); err != nil {
			return nil, fmt.Errorf("error unmarshaling session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s *SessionStorage) GetSessionsForModel(modelID string) ([]string, error) {
	keys, err := s.db.GetPrefix(formatModelSessionKey(modelID, ""))
	if err != nil {
		return []string{}, err
	}

	sessionIDs := make([]string, len(keys))
	for i, key := range keys {
		_, sessionID := parseModelSessionKey(key)
		sessionIDs[i] = sessionID
	}

	return sessionIDs, nil
}

func formatActivityPrefix(modelID string) []byte {
	return []byte(fmt.Sprintf("activity:%s:", strings.ToLower(modelID)))
}

func formatActivityKey(modelID string, endTime int64, sessionID string) []byte {
	return []byte(fmt.Sprintf("activity:%s:%d:%s", strings.ToLower(modelID), endTime, strings.ToLower(sessionID)))
}

func parseActivityKeyTimestamp(key []byte) (int64, error) {
	parts := strings.Split(string(key), ":")
	if len(parts) < 4 {
		return 0, fmt.Errorf("invalid activity key format: %s", string(key))
	}
	return strconv.ParseInt(parts[2], 10, 64)
}

// AddActivity stores each activity as an individual key-value entry.
func (s *SessionStorage) AddActivity(modelID string, activity *PromptActivity) error {
	modelID = strings.ToLower(modelID)
	key := formatActivityKey(modelID, activity.EndTime, activity.SessionID)
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		return err
	}
	return s.db.SetWithTTL(key, activityJSON, ActivityTTL)
}

func (s *SessionStorage) GetActivities(modelID string) ([]*PromptActivity, error) {
	modelID = strings.ToLower(modelID)
	_, values, err := s.db.GetPrefixWithValues(formatActivityPrefix(modelID))
	if err != nil {
		return nil, fmt.Errorf("error reading activities for model %s: %w", modelID, err)
	}

	activities := make([]*PromptActivity, 0, len(values))
	for _, activityJSON := range values {
		var activity PromptActivity
		if err := json.Unmarshal(activityJSON, &activity); err != nil {
			return nil, fmt.Errorf("error unmarshaling activity for model %s: %w", modelID, err)
		}
		activities = append(activities, &activity)
	}

	return activities, nil
}

// RemoveOldActivities removes individual activity keys older than beforeTime.
func (s *SessionStorage) RemoveOldActivities(modelID string, beforeTime int64) error {
	modelID = strings.ToLower(modelID)
	keys, err := s.db.GetPrefix(formatActivityPrefix(modelID))
	if err != nil {
		return fmt.Errorf("error reading activity keys for model %s: %w", modelID, err)
	}

	return s.db.RunInTransaction(func(txn *badger.Txn) error {
		for _, key := range keys {
			endTime, err := parseActivityKeyTimestamp(key)
			if err != nil {
				continue
			}
			if endTime <= beforeTime {
				if err := txn.Delete(key); err != nil {
					return fmt.Errorf("error deleting activity key %s: %w", string(key), err)
				}
			}
		}
		return nil
	})
}

func formatModelSessionKey(modelID string, sessionID string) []byte {
	return []byte(fmt.Sprintf("model:%s:session:%s", strings.ToLower(modelID), strings.ToLower(sessionID)))
}

func formatSessionKey(sessionID string) []byte {
	return []byte(fmt.Sprintf("session:%s", strings.ToLower(sessionID)))
}

func parseModelSessionKey(key []byte) (string, string) {
	parts := strings.Split(string(key), ":")
	return parts[1], parts[3]
}

func parseSessionKey(key []byte) (string, string) {
	parts := strings.Split(string(key), ":")
	return parts[0], parts[1]
}
