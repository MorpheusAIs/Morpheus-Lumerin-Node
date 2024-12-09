package proxyapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	msg "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

type ProxyReceiver struct {
	privateKeyHex     lib.HexString
	publicKeyHex      lib.HexString
	chainID           *big.Int
	morRpc            *m.MORRPCMessage
	sessionStorage    *storages.SessionStorage
	aiEngine          *aiengine.AiEngine
	modelConfigLoader *config.ModelConfigLoader
	service           BidGetter
	sessionRepo       *sessionrepo.SessionRepositoryCached

	useDhEncryption bool
	dhKeysMap       map[string]string
}

func NewProxyReceiver(privateKeyHex, publicKeyHex lib.HexString, sessionStorage *storages.SessionStorage, aiEngine *aiengine.AiEngine, chainID *big.Int, modelConfigLoader *config.ModelConfigLoader, blockchainService BidGetter, sessionRepo *sessionrepo.SessionRepositoryCached, useDhEncryption bool) *ProxyReceiver {
	return &ProxyReceiver{
		privateKeyHex:     privateKeyHex,
		publicKeyHex:      publicKeyHex,
		morRpc:            m.NewMorRpc(),
		aiEngine:          aiEngine,
		chainID:           chainID,
		modelConfigLoader: modelConfigLoader,
		service:           blockchainService,
		sessionStorage:    sessionStorage,
		sessionRepo:       sessionRepo,
		dhKeysMap:         make(map[string]string),
		useDhEncryption:   useDhEncryption,
	}
}

func (s *ProxyReceiver) SessionPrompt(ctx context.Context, requestID string, userPubKey string, rq *m.SessionPromptReq, sendResponse SendResponse, sourceLog lib.ILogger) (int, int, error) {
	session, err := s.sessionRepo.GetSession(ctx, rq.SessionID)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get session"), err)
		sourceLog.Error(err)
		return 0, 0, err
	}

	var req *openai.ChatCompletionRequest
	sharedEncryptionKey := []byte{}
	if s.useDhEncryption {
		key := fmt.Sprintf("shared_key_%s", session.UserAddr())
		sharedEncryptionKey = []byte(s.dhKeysMap[key])

		promptBytes := common.FromHex(rq.Message)
		dectyptedMessage, err := lib.SharedSecretDecrypt(sharedEncryptionKey, promptBytes)
		if err != nil {
			return 0, 0, err
		}

		err = json.Unmarshal(dectyptedMessage, &req)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to unmarshal prompt"), err)
			sourceLog.Error(err)
			return 0, 0, err
		}
	} else {
		promptBytes := common.FromHex(rq.Message)
		err := json.Unmarshal(promptBytes, &req)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to unmarshal prompt"), err)
			sourceLog.Error(err)
			return 0, 0, err
		}
	}

	ttftMs := 0
	totalTokens := 0
	now := time.Now().UnixMilli()

	adapter, err := s.aiEngine.GetAdapter(ctx, common.Hash{}, session.ModelID(), common.Hash{}, false, false)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get adapter"), err)
		sourceLog.Error(err)
		return 0, 0, err
	}

	err = adapter.Prompt(ctx, req, func(ctx context.Context, completion genericchatstorage.Chunk) error {
		totalTokens += completion.Tokens()

		if ttftMs == 0 {
			ttftMs = int(time.Now().UnixMilli() - now)
		}

		marshalledResponse, err := json.Marshal(completion.Data())
		if err != nil {
			return err
		}

		var encryptedResponse []byte
		if s.useDhEncryption {
			encryptedResponse, err = lib.SharedSecretEncrypt([]byte(sharedEncryptionKey), marshalledResponse)
			if err != nil {
				return err
			}
		} else {
			encryptedResponseStr, err := lib.EncryptString(string(marshalledResponse), lib.RemoveHexPrefix(userPubKey))
			if err != nil {
				return err
			}
			encryptedResponse, err = hex.DecodeString(encryptedResponseStr)
			if err != nil {
				return err
			}
		}
		responseHex := lib.HexString(encryptedResponse)

		// Send response
		r, err := s.morRpc.SessionPromptResponse(
			responseHex.Hex(),
			s.privateKeyHex,
			requestID,
		)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to create response"), err)
			sourceLog.Error(err)
			return err
		}
		return sendResponse(r)
	})
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to prompt"), err)
		sourceLog.Error(err)
		return 0, 0, err
	}

	activity := storages.PromptActivity{
		SessionID: session.ID().Hex(),
		StartTime: now,
		EndTime:   time.Now().Unix(),
	}
	err = s.sessionStorage.AddActivity(session.ModelID().Hex(), &activity)
	if err != nil {
		sourceLog.Warnf("failed to store activity: %s", err)
	}

	return ttftMs, totalTokens, nil
}

func (s *ProxyReceiver) SessionRequest(ctx context.Context, msgID string, reqID string, req *m.SessionReq, log lib.ILogger) (*msg.RpcResponse, error) {
	log.Debugf("Received session request from %s, timestamp: %s", req.User, req.Timestamp)

	bid, err := s.service.GetBidByID(ctx, req.BidID)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get bid"), err)
		log.Error(err)
		return nil, err
	}

	modelID := bid.ModelAgentId.String()
	modelConfig := s.modelConfigLoader.ModelConfigFromID(modelID)
	capacityManager := CreateCapacityManager(modelConfig, s.sessionStorage, log)

	hasCapacity := capacityManager.HasCapacity(modelID)
	if !hasCapacity {
		err := fmt.Errorf("no capacity")
		log.Error(err)
		return nil, err
	}

	providerPublicKeyForSharedSecret := []byte{}
	if s.useDhEncryption {
		pubKey, providerPrivateKey, err := lib.GenerateEphemeralKeyPair()
		if err != nil {
			return nil, err
		}

		s.dhKeysMap[req.User.Hex()] = string(providerPrivateKey)
		providerPublicKeyForSharedSecret = pubKey
	}

	// Send response
	response, err := s.morRpc.InitiateSessionResponse(
		s.publicKeyHex,
		req.User,
		req.BidID,
		s.privateKeyHex,
		reqID,
		s.chainID,
		providerPublicKeyForSharedSecret,
	)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to create response"), err)
		log.Error(err)
		return nil, err
	}

	user := storages.User{
		Addr:   req.User.Hex(),
		PubKey: req.Key.Hex(),
	}

	err = s.sessionStorage.AddUser(&user)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed store user"), err)
		log.Error(err)
		return nil, err
	}

	return response, nil
}

func (s *ProxyReceiver) SessionReport(ctx context.Context, msgID string, reqID string, session *storages.Session, sourceLog lib.ILogger) (*msg.RpcResponse, error) {
	sourceLog.Debugf("received session report request for %s", session.Id)

	tps := 0
	ttft := 0
	for _, tpsVal := range session.TPSScaled1000Arr {
		tps += tpsVal
	}
	for _, ttftVal := range session.TTFTMsArr {
		ttft += ttftVal
	}

	if len(session.TPSScaled1000Arr) != 0 {
		tps /= len(session.TPSScaled1000Arr)
	}
	if len(session.TTFTMsArr) != 0 {
		ttft /= len(session.TTFTMsArr)
	}

	response, err := s.morRpc.SessionReportResponse(
		uint32(tps),
		uint32(ttft),
		common.HexToHash(session.Id),
		s.privateKeyHex,
		reqID,
		s.chainID,
	)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to create response"), err)
		sourceLog.Error(err)
		return nil, err
	}

	return response, nil
}

func (s *ProxyReceiver) CreateSharedEncryptionKey(ctx context.Context, msgID string, reqID string, data *m.CreateSharedEncrKeyReq, sourceLog lib.ILogger) error {
	providerPrivateKey := s.dhKeysMap[data.UserAddress.String()]

	sharedSecret, err := lib.ComputeSharedSecret([]byte(providerPrivateKey), data.UserPublicKeyForSharedSecret)
	if err != nil {
		return err
	}

	encryptionKey, err := lib.DeriveKeysFromSharedSecret(sharedSecret)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("shared_key_%s", data.UserAddress.String())
	s.dhKeysMap[key] = string(encryptionKey)

	return nil
}
