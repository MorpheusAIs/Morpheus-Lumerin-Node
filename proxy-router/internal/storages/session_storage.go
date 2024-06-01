package storages

type Session struct {
	Id           string
	UserAddr     string
	ProviderAddr string
}

type User struct {
	Addr   string
	PubKey string
}

type SessionStorage struct {
	storage map[string]*Session
	users   map[string]*User
}

func NewSessionStorage() *SessionStorage {
	return &SessionStorage{
		storage: make(map[string]*Session),
		users:   make(map[string]*User),
	}
}

func (s *SessionStorage) GetSession(id string) (*Session, bool) {
	session, ok := s.storage[id]
	return session, ok
}

func (s *SessionStorage) GetUser(addr string) (*User, bool) {
	user, ok := s.users[addr]
	return user, ok
}

func (s *SessionStorage) AddSession(session *Session) {
	s.storage[session.Id] = session
}

func (s *SessionStorage) AddUser(user *User) {
	s.users[user.Addr] = user
}

func (s *SessionStorage) RemoveSession(id string) {
	delete(s.storage, id)
}
