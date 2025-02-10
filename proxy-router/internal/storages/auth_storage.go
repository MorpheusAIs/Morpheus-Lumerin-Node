package storages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const (
	authRequestPrefix      = "auth_request"
	allowanceRequestPrefix = "allowance_request"
	agentTxPrefix          = "agent_tx"
)

type AuthStorage struct {
	db *Storage
}

func NewAuthStorage(storage *Storage) *AuthStorage {
	return &AuthStorage{
		db: storage,
	}
}

func (s *AuthStorage) AddAuthRequest(request *AgentUser) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	key := formatKey(authRequestPrefix, request.Username)
	err = s.db.Set(key, requestJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthStorage) GetAgentUser(username string) (*AgentUser, bool) {
	key := formatKey(authRequestPrefix, username)
	requestJson, err := s.db.Get(key)
	if err != nil {
		return nil, false
	}

	request := &AgentUser{}
	err = json.Unmarshal(requestJson, request)
	if err != nil {
		return nil, false
	}

	return request, true
}

func (s *AuthStorage) GetAgentUsers() ([]*AgentUser, error) {
	var requests []*AgentUser

	prefix := formatPrefix(authRequestPrefix)
	keys, err := s.db.GetPrefix(prefix)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		username := trimPrefix(key, prefix)
		request, ok := s.GetAgentUser(string(username))
		if !ok {
			return nil, fmt.Errorf("error getting auth request: %s", string(key))
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *AuthStorage) DeleteAuthRequest(username string) error {
	key := formatKey(authRequestPrefix, username)
	return s.db.Delete(key)
}

func (s *AuthStorage) AddAllowanceRequest(request *AllowanceRequest) error {
	key := formatKey(allowanceRequestPrefix, request.Username, request.Token)
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return s.db.Set(key, requestJson)
}

func (s *AuthStorage) GetAllowanceRequest(username string, token string) (*AllowanceRequest, bool) {
	token = strings.ToLower(token)
	key := formatKey(allowanceRequestPrefix, username, token)
	requestJson, err := s.db.Get(key)
	if err != nil {
		return nil, false
	}

	request := &AllowanceRequest{}
	err = json.Unmarshal(requestJson, request)
	if err != nil {
		return nil, false
	}

	return request, true
}

func (s *AuthStorage) GetAllowanceRequests() ([]*AllowanceRequest, error) {
	var requests []*AllowanceRequest

	prefix := formatPrefix(allowanceRequestPrefix)
	keys, err := s.db.GetPrefix(prefix)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		// Split key into username and token
		parts := bytes.Split(trimPrefix(key, prefix), []byte(":"))
		if len(parts) != 2 {
			continue
		}

		username, token := string(parts[0]), string(parts[1])
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
	key := formatKey(allowanceRequestPrefix, username, token)
	return s.db.Delete(key)
}

func (s *AuthStorage) SetAllowance(username string, token string, amount lib.BigInt) error {
	key := formatKey(authRequestPrefix, username)
	requestJson, err := s.db.Get(key)
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

	return s.db.Set(key, updatedJson)
}

func (s *AuthStorage) SetAgentTx(txHash string, username string, blockNumber *big.Int) error {
	// reversing to maintain reverse order of indexing
	reversedBlockNumber := math.MaxUint64 - blockNumber.Uint64()
	key := formatKey(agentTxPrefix, username, strconv.FormatUint(reversedBlockNumber, 10))
	return s.db.Set(key, []byte(txHash))
}

func (s *AuthStorage) GetAgentTxs(username string, cursor []byte, limit uint) ([]string, []byte, error) {
	txs := make([]string, 0)
	prefix := formatPrefix(agentTxPrefix, username)

	keys, nextCursor, err := s.db.Paginate(prefix, cursor, limit)
	if err != nil {
		return nil, nil, err
	}

	for _, key := range keys {
		txhash, err := s.db.Get(key)
		if err != nil {
			return nil, nil, err
		}
		txs = append(txs, string(txhash))
	}
	return txs, nextCursor, nil
}

// formatKey formats a key by joining the path components with a colon
func formatKey(path ...string) []byte {
	return []byte(strings.Join(path, ":"))
}

// formatPrefix formats a prefix by joining the path components with a colon and adding a trailing colon
func formatPrefix(path ...string) []byte {
	return []byte(strings.Join(path, ":") + ":")
}

func trimPrefix(key []byte, prefix []byte) []byte {
	return bytes.TrimPrefix(key, prefix)
}
