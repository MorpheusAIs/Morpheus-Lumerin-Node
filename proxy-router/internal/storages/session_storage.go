package storages

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SessionStorage struct {
	db *Storage
}

func NewSessionStorage(storage *Storage) *SessionStorage {
	return &SessionStorage{
		db: storage,
	}
}

func (s *SessionStorage) GetUser(addr string) (*User, bool) {
	addr = strings.ToLower(addr)
	key := fmt.Sprintf("user:%s", addr)
	userJson, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, false
	}

	user := &User{}
	err = json.Unmarshal(userJson, user)
	if err != nil {
		return nil, false
	}

	return user, true
}

func (s *SessionStorage) AddUser(user *User) error {
	addr := strings.ToLower(user.Addr)
	key := fmt.Sprintf("user:%s", addr)
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = s.db.Set([]byte(key), userJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionStorage) GetSession(id string) (*Session, bool) {
	sessionJson, err := s.db.Get(formatSessionKey(id))
	if err != nil {
		return nil, false
	}

	session := &Session{}
	err = json.Unmarshal(sessionJson, session)
	if err != nil {
		return nil, false
	}

	return session, true
}

func (s *SessionStorage) AddSession(session *Session) error {
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// TODO: do in a single transaction
	err = s.db.Set(formatSessionKey(session.Id), sessionJson)
	if err != nil {
		return err
	}
	err = s.addSessionToModel(session.ModelID, session.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionStorage) RemoveSession(id string) error {
	ses, ok := s.GetSession(id)
	if !ok {
		return nil
	}
	sessionId := strings.ToLower(id)
	key := fmt.Sprintf("session:%s", sessionId)

	err := s.db.Delete([]byte(key))
	if err != nil {
		return err
	}

	return s.removeSessionFromModel(ses.ModelID, sessionId)
}

func (s *SessionStorage) addSessionToModel(modelID string, sessionID string) error {
	err := s.db.Set(formatModelSessionKey(modelID, sessionID), []byte{})
	if err != nil {
		return fmt.Errorf("error adding session to model: %s", err)
	}
	return nil
}

func (s *SessionStorage) removeSessionFromModel(modelID string, sessionID string) error {
	err := s.db.Delete(formatModelSessionKey(modelID, sessionID))
	if err != nil {
		return fmt.Errorf("error removing session from model: %s", err)
	}
	return nil
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

func (s *SessionStorage) AddActivity(modelID string, activity *PromptActivity) error {
	modelID = strings.ToLower(modelID)
	key := fmt.Sprintf("activity:%s", modelID)

	var activities []*PromptActivity
	activitiesJson, err := s.db.Get([]byte(key))
	if err == nil {
		err = json.Unmarshal(activitiesJson, &activities)
		if err != nil {
			return err
		}
	} else {
		activities = []*PromptActivity{}
	}

	activities = append(activities, activity)
	activitiesJson, err = json.Marshal(activities)
	if err != nil {
		return err
	}

	err = s.db.Set([]byte(key), activitiesJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionStorage) GetActivities(modelID string) ([]*PromptActivity, error) {
	modelID = strings.ToLower(modelID)
	key := fmt.Sprintf("activity:%s", modelID)

	activitiesJson, err := s.db.Get([]byte(key))
	if err != nil {
		return []*PromptActivity{}, nil
	}

	var activities []*PromptActivity
	err = json.Unmarshal(activitiesJson, &activities)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

// // New method to remove activities older than a certain time
func (s *SessionStorage) RemoveOldActivities(modelID string, beforeTime int64) error {
	modelID = strings.ToLower(modelID)
	key := fmt.Sprintf("activity:%s", modelID)

	activitiesJson, err := s.db.Get([]byte(key))
	if err != nil {
		return nil
	}

	var activities []*PromptActivity
	err = json.Unmarshal(activitiesJson, &activities)
	if err != nil {
		return err
	}

	// Filter activities, keep only those after beforeTime
	var updatedActivities []*PromptActivity
	for _, activity := range activities {
		if activity.EndTime > beforeTime {
			updatedActivities = append(updatedActivities, activity)
		}
	}

	activitiesJson, err = json.Marshal(updatedActivities)
	if err != nil {
		return err
	}

	err = s.db.Set([]byte(key), activitiesJson)
	if err != nil {
		return err
	}

	return nil
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
