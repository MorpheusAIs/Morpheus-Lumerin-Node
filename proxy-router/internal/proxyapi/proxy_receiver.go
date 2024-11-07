package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
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
}

func NewProxyReceiver(privateKeyHex, publicKeyHex lib.HexString, sessionStorage *storages.SessionStorage, aiEngine *aiengine.AiEngine, chainID *big.Int, modelConfigLoader *config.ModelConfigLoader, blockchainService BidGetter, sessionRepo *sessionrepo.SessionRepositoryCached) *ProxyReceiver {
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
	}
}

func (s *ProxyReceiver) SessionPrompt(ctx context.Context, requestID string, userPubKey string, rq *m.SessionPromptReq, sendResponse SendResponse, sourceLog lib.ILogger) (int, int, error) {
	var req *openai.ChatCompletionRequest

	err := json.Unmarshal([]byte(rq.Message), &req)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to unmarshal prompt"), err)
		sourceLog.Error(err)
		return 0, 0, err
	}

	session, err := s.sessionRepo.GetSession(ctx, rq.SessionID)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get session"), err)
		sourceLog.Error(err)
		return 0, 0, err
	}

	ttftMs := 0
	totalTokens := 0
	now := time.Now().UnixMilli()

	responseCb := func(response interface{}) error {
		openAiResponse, ok := response.(*openai.ChatCompletionStreamResponse)
		if ok {
			totalTokens += len(openAiResponse.Choices)
		} else {
			_, ok := response.(*aiengine.ProdiaGenerationResult)
			if ok {
				totalTokens += 1
			} else {
				return fmt.Errorf("unknown response type")
			}
		}

		if ttftMs == 0 {
			ttftMs = int(time.Now().UnixMilli() - now)
		}

		marshalledResponse, err := json.Marshal(response)
		if err != nil {
			return err
		}

		encryptedResponse, err := lib.EncryptString(string(marshalledResponse), lib.RemoveHexPrefix(userPubKey))
		if err != nil {
			return err
		}

		// Send response
		r, err := s.morRpc.SessionPromptResponse(
			encryptedResponse,
			s.privateKeyHex,
			requestID,
		)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to create response"), err)
			sourceLog.Error(err)
			return err
		}
		return sendResponse(r)
	}

	modelConfig := s.modelConfigLoader.ModelConfigFromID(session.ModelID().Hex())
	if modelConfig == nil {
		return 0, 0, fmt.Errorf("model config not found for model id %s", session.ModelID())
	}

	if modelConfig.ApiType == "prodia" {
		lastMessage := req.Messages[len(req.Messages)-1]
		prodiaReq := &aiengine.ProdiaGenerationRequest{
			Prompt: lastMessage.Content,
			Model:  modelConfig.ModelName,
			ApiUrl: modelConfig.ApiURL,
			ApiKey: modelConfig.ApiKey,
		}

		err = s.aiEngine.PromptProdiaImage(ctx, prodiaReq, responseCb)
	} else {
		req.Model = modelConfig.ModelName
		if req.Model == "" {
			req.Model = "llama2"
		}
		_, err = s.aiEngine.PromptStream(ctx, req, responseCb)
	}

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

	// Send response
	response, err := s.morRpc.InitiateSessionResponse(
		s.publicKeyHex,
		req.User,
		req.BidID,
		s.privateKeyHex,
		reqID,
		s.chainID,
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
	sourceLog.Debugf("Received session report request for %s, timestamp: %s", session.Id)

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
