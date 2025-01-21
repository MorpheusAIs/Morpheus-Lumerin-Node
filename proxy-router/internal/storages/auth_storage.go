package storages

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type AgentUser struct {
	Username    string
	Password    string
	Perms       []string
	Allowances  map[string]lib.BigInt
	IsConfirmed bool
}

type AllowanceRequest struct {
	Username  string
	Token     string
	Allowance lib.BigInt
}

type AuthStorage struct {
	db *Storage
}

func NewAuthStorage(storage *Storage) *AuthStorage {
	return &AuthStorage{
		db: storage,
	}
}

func (s *AuthStorage) AddAuthRequest(request *AgentUser) error {
	key := fmt.Sprintf("auth_request:%s", request.Username)
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = s.db.Set([]byte(key), requestJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthStorage) GetAgentUser(username string) (*AgentUser, bool) {
	key := fmt.Sprintf("auth_request:%s", username)
	requestJson, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, false
	}

	request := &AgentUser{}
	json.Unmarshal(requestJson, request)
	return request, true
}

func (s *AuthStorage) GetAgentUsers() ([]*AgentUser, error) {
	var requests []*AgentUser
	prefix := []byte("auth_request:")

	keys, err := s.db.GetPrefix(prefix)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		username := strings.TrimPrefix(string(key), "auth_request:")
		request, ok := s.GetAgentUser(username)
		if !ok {
			return nil, fmt.Errorf("error getting auth request: %s", string(key))
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *AuthStorage) DeleteAuthRequest(username string) error {
	key := fmt.Sprintf("auth_request:%s", username)
	return s.db.Delete([]byte(key))
}

func (s *AuthStorage) AddAllowanceRequest(request *AllowanceRequest) error {
	key := fmt.Sprintf("allowance_request:%s:%s", request.Username, request.Token)
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return s.db.Set([]byte(key), requestJson)
}

func (s *AuthStorage) GetAllowanceRequest(username string, token string) (*AllowanceRequest, bool) {
	token = strings.ToLower(token)
	key := fmt.Sprintf("allowance_request:%s:%s", username, token)
	requestJson, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, false
	}

	request := &AllowanceRequest{}
	json.Unmarshal(requestJson, request)
	return request, true
}

func (s *AuthStorage) GetAllowanceRequests() ([]*AllowanceRequest, error) {
	var requests []*AllowanceRequest
	prefix := []byte("allowance_request:")

	keys, err := s.db.GetPrefix(prefix)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		// Split key into username and token
		parts := strings.Split(strings.TrimPrefix(string(key), "allowance_request:"), ":")
		if len(parts) != 2 {
			continue
		}
		username, token := parts[0], parts[1]

		request, ok := s.GetAllowanceRequest(username, token)
		if !ok {
			return nil, fmt.Errorf("error getting allowance request: %s", string(key))
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *AuthStorage) ConfirmOrDeclineAllowanceRequest(username string, token string, isConfirmed bool) error {
	token = strings.ToLower(token)
	request, ok := s.GetAllowanceRequest(username, token)
	if !ok {
		return fmt.Errorf("allowance request not found for user %s and token %s", username, token)
	}

	if isConfirmed {
		err := s.SetAllowance(username, token, request.Allowance)
		if err != nil {
			return err
		}
	}

	// Delete the request after processing (whether confirmed or declined)
	key := fmt.Sprintf("allowance_request:%s:%s", username, token)
	return s.db.Delete([]byte(key))
}

func (s *AuthStorage) SetAllowance(username string, token string, amount lib.BigInt) error {
	key := fmt.Sprintf("auth_request:%s", username)
	requestJson, err := s.db.Get([]byte(key))
	if err != nil {
		return err
	}

	request := &AgentUser{}
	err = json.Unmarshal(requestJson, request)
	if err != nil {
		return err
	}

	if request.Allowances == nil {
		request.Allowances = make(map[string]lib.BigInt)
	}

	request.Allowances[token] = amount
	updatedJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return s.db.Set([]byte(key), updatedJson)
}

func (s *AuthStorage) SetAgentTx(txHash string, username string) error {
	key := fmt.Sprintf("agent_tx:%s", txHash)
	return s.db.Set([]byte(key), []byte(username))
}

func (s *AuthStorage) GetAgentTxs() (map[string]string, error) {
	var txs map[string]string = make(map[string]string)
	prefix := []byte("agent_tx:")

	keys, err := s.db.GetPrefix(prefix)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		txHash := strings.TrimPrefix(string(key), "agent_tx:")
		username, err := s.db.Get([]byte(key))
		if err != nil {
			return nil, err
		}
		txs[txHash] = string(username)
	}
	return txs, nil
}
