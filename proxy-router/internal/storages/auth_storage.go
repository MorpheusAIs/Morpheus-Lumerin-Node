package storages

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	badger "github.com/dgraph-io/badger/v4"
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
	return s.db.Set(key, requestJson)
}

// GetAgentUser retrieves an agent user by username. Returns (nil, nil) if not found.
// Returns a non-nil error only on actual storage or deserialization failures.
func (s *AuthStorage) GetAgentUser(username string) (*AgentUser, error) {
	key := formatKey(authRequestPrefix, username)
	requestJson, err := s.db.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading agent user %s: %w", username, err)
	}

	request := &AgentUser{}
	if err := json.Unmarshal(requestJson, request); err != nil {
		return nil, fmt.Errorf("error unmarshaling agent user %s: %w", username, err)
	}

	return request, nil
}

func (s *AuthStorage) GetAgentUsers() ([]*AgentUser, error) {
	prefix := formatPrefix(authRequestPrefix)
	_, values, err := s.db.GetPrefixWithValues(prefix)
	if err != nil {
		return nil, fmt.Errorf("error reading agent users: %w", err)
	}

	requests := make([]*AgentUser, 0, len(values))
	for _, val := range values {
		request := &AgentUser{}
		if err := json.Unmarshal(val, request); err != nil {
			return nil, fmt.Errorf("error unmarshaling agent user: %w", err)
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

// GetAllowanceRequest retrieves an allowance request. Returns (nil, nil) if not found.
// Returns a non-nil error only on actual storage or deserialization failures.
func (s *AuthStorage) GetAllowanceRequest(username string, token string) (*AllowanceRequest, error) {
	token = strings.ToLower(token)
	key := formatKey(allowanceRequestPrefix, username, token)
	requestJson, err := s.db.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading allowance request for %s/%s: %w", username, token, err)
	}

	request := &AllowanceRequest{}
	if err := json.Unmarshal(requestJson, request); err != nil {
		return nil, fmt.Errorf("error unmarshaling allowance request for %s/%s: %w", username, token, err)
	}

	return request, nil
}

func (s *AuthStorage) GetAllowanceRequests() ([]*AllowanceRequest, error) {
	prefix := formatPrefix(allowanceRequestPrefix)
	keys, values, err := s.db.GetPrefixWithValues(prefix)
	if err != nil {
		return nil, fmt.Errorf("error reading allowance requests: %w", err)
	}

	requests := make([]*AllowanceRequest, 0, len(values))
	for i, val := range values {
		// Validate key structure
		parts := bytes.Split(trimPrefix(keys[i], prefix), []byte(":"))
		if len(parts) != 2 {
			continue
		}

		request := &AllowanceRequest{}
		if err := json.Unmarshal(val, request); err != nil {
			return nil, fmt.Errorf("error unmarshaling allowance request %s: %w", string(keys[i]), err)
		}
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *AuthStorage) ConfirmOrDeclineAllowanceRequest(username string, token string, isConfirmed bool) error {
	token = strings.ToLower(token)
	request, err := s.GetAllowanceRequest(username, token)
	if err != nil {
		return fmt.Errorf("error looking up allowance request: %w", err)
	}
	if request == nil {
		return fmt.Errorf("allowance request not found for user %s and token %s", username, token)
	}

	if isConfirmed {
		if err := s.SetAllowance(username, token, request.Allowance); err != nil {
			return err
		}
	}

	key := formatKey(allowanceRequestPrefix, username, token)
	return s.db.Delete(key)
}

// SetAllowance atomically reads the agent user, updates the allowance, and writes back
// in a single BadgerDB transaction to prevent race conditions.
func (s *AuthStorage) SetAllowance(username string, token string, amount lib.BigInt) error {
	key := formatKey(authRequestPrefix, username)

	return s.db.RunInTransaction(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("error reading agent user %s for allowance update: %w", username, err)
		}

		requestJson, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("error copying agent user value for %s: %w", username, err)
		}

		request := &AgentUser{}
		if err := json.Unmarshal(requestJson, request); err != nil {
			return fmt.Errorf("error unmarshaling agent user %s: %w", username, err)
		}

		if request.Allowances == nil {
			request.Allowances = make(map[string]lib.BigInt)
		}

		request.Allowances[token] = amount
		updatedJson, err := json.Marshal(request)
		if err != nil {
			return err
		}

		return txn.Set(key, updatedJson)
	})
}

func (s *AuthStorage) SetAgentTx(txHash string, username string, blockNumber *big.Int) error {
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
			return nil, nil, fmt.Errorf("error reading tx for key %s: %w", string(key), err)
		}
		txs = append(txs, string(txhash))
	}
	return txs, nextCursor, nil
}

func formatKey(path ...string) []byte {
	return []byte(strings.Join(path, ":"))
}

func formatPrefix(path ...string) []byte {
	return []byte(strings.Join(path, ":") + ":")
}

func trimPrefix(key []byte, prefix []byte) []byte {
	return bytes.TrimPrefix(key, prefix)
}
