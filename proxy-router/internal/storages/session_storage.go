package storages

type Session struct {
	Id             string
	UserPubKey     string
	ProviderPubKey string
}

type SessionStorage struct {
	storage map[string]*Session
}

func NewSessionStorage() *SessionStorage {
	return &SessionStorage{
		storage: make(map[string]*Session),
	}
}

func (s *SessionStorage) GetSession(id string) *Session {
	return s.storage[id]
}

func (s *SessionStorage) AddSession(session *Session) {
	s.storage[session.Id] = session
}

func (s *SessionStorage) RemoveSession(id string) {
	delete(s.storage, id)
}
