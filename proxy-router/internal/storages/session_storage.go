package storages

import (
	"encoding/json"
	"fmt"

	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
)

type Session struct {
	Id           string
	UserAddr     string
	ProviderAddr string
}

type User struct {
	Addr   string
	PubKey string
	Url    string
}

type SessionStorage struct {
	db *Storage
}

func NewSessionStorage(log i.ILogger) *SessionStorage {
	return &SessionStorage{
		db: NewStorage(log),
	}
}

func (s *SessionStorage) GetSession(id string) (*Session, bool) {
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
	key := fmt.Sprintf("session:%s", session.Id)
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
	key := fmt.Sprintf("user:%s", user.Addr)
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
	key := fmt.Sprintf("session:%s", id)
	err := s.db.Delete([]byte(key))
	if err != nil {
		return err
	}
	return nil
}
