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

func (s *SessionStorage) GetSession(id string) (*Session, bool) {
	id = strings.ToLower(id)
	key := fmt.Sprintf("session:%s", id)

	sessionJson, err := s.db.Get([]byte(key))
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

func (s *SessionStorage) AddSession(session *Session) error {
	sessionId := strings.ToLower(session.Id)
	key := fmt.Sprintf("session:%s", sessionId)
	sessionJson, err := json.Marshal(session)
	if err != nil {
		return err
	}

	err = s.db.Set([]byte(key), sessionJson)
	if err != nil {
		return err
	}
	return nil
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

func (s *SessionStorage) RemoveSession(id string) error {
	sessionId := strings.ToLower(id)
	key := fmt.Sprintf("session:%s", sessionId)
	err := s.db.Delete([]byte(key))
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionStorage) AddSessionToModel(modelID string, sessionID string) error {
	modelID = strings.ToLower(modelID)
	key := fmt.Sprintf("model:%s", modelID)

	var sessions []string
	sessionsJson, err := s.db.Get([]byte(key))
	if err == nil {
		err = json.Unmarshal(sessionsJson, &sessions)
		if err != nil {
			return err
		}
	} else {
		sessions = []string{}
	}

	sessions = append(sessions, sessionID)
	sessionsJson, err = json.Marshal(sessions)
	if err != nil {
		return err
	}

	err = s.db.Set([]byte(key), sessionsJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionStorage) GetSessionsForModel(modelID string) ([]string, error) {
	modelID = strings.ToLower(modelID)
	key := fmt.Sprintf("model:%s", modelID)

	sessionsJson, err := s.db.Get([]byte(key))
	if err != nil {
		return []string{}, nil
	}

	var sessions []string
	err = json.Unmarshal(sessionsJson, &sessions)
	if err != nil {
		return nil, err
	}

	return sessions, nil
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
