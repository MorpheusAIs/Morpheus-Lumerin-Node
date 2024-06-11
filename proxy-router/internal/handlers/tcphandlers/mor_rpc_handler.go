package tcphandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/apibus"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/morrpc"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/storages"
	"github.com/sashabaranov/go-openai"
)

type MorRpcHandler struct {
	privateKeyHex  string
	publicKeyHex   string
	address        string
	morRpc         *morrpc.MorRpc
	sessionStorage *storages.SessionStorage
	apibus         *apibus.ApiBus
}

func NewMorRpcHandler(privateKeyHex string, morRpc *morrpc.MorRpc, sessionStorage *storages.SessionStorage, apiBus *apibus.ApiBus) *MorRpcHandler {
	return &MorRpcHandler{
		privateKeyHex:  privateKeyHex,
		address:        lib.MustPrivKeyStringToAddr(privateKeyHex).Hex(),
		publicKeyHex:   lib.MustPubKeyStringFromPrivate(privateKeyHex),
		morRpc:         morRpc,
		sessionStorage: sessionStorage,
		apibus:         apiBus,
	}
}

func (m *MorRpcHandler) Handle(ctx context.Context, msg morrpc.RpcMessage, sourceLog interfaces.ILogger, sendCallback func(*morrpc.RpcResponse) error) error {
	switch msg.Method {
	case "session.request":
		requestId := fmt.Sprintf("%v", msg.ID)
		signature := fmt.Sprintf("%v", msg.Params["signature"])
		userAddr := fmt.Sprintf("%v", msg.Params["user"])
		userPubKey := fmt.Sprintf("%v", msg.Params["key"])
		spend := fmt.Sprintf("%v", msg.Params["spend"])
		timeStamp := fmt.Sprintf("%v", msg.Params["timestamp"])
		bidId := fmt.Sprintf("%v", msg.Params["bidid"])
		sourceLog.Debugf("Received session request from %s, timestamp: %s", userAddr, timeStamp)

		isValid := m.morRpc.VerifySignature(msg.Params, signature, userPubKey, sourceLog)
		if !isValid {
			err := fmt.Errorf("invalid signature")
			sourceLog.Error(err)
			return err
		}

		hasCapacity := true // check if there is capacity
		if !hasCapacity && spend != "" {
			err := fmt.Errorf("no capacity")
			sourceLog.Error(err)
			return err
		}

		// Send response
		response, err := m.morRpc.InitiateSessionResponse(
			m.publicKeyHex,
			userAddr,
			bidId,
			m.privateKeyHex,
			requestId,
		)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to create response"), err)
			sourceLog.Error(err)
			return err
		}

		user := storages.User{
			Addr:   userAddr,
			PubKey: userPubKey,
		}

		err = m.sessionStorage.AddUser(&user)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed store user"), err)
			sourceLog.Error(err)
			return err
		}
		sendCallback(response)
		return nil
	case "session.prompt":
		requestId := fmt.Sprintf("%v", msg.ID)
		signature := fmt.Sprintf("%v", msg.Params["signature"])
		sessionId := fmt.Sprintf("%v", msg.Params["sessionid"])
		prompt := fmt.Sprintf("%v", msg.Params["message"])
		timeStamp := fmt.Sprintf("%v", msg.Params["timestamp"])
		sourceLog.Debugf("Received prompt from session %s, timestamp: %s", sessionId, timeStamp)
		session, ok := m.sessionStorage.GetSession(sessionId)
		if !ok {
			err := fmt.Errorf("session not found")
			sourceLog.Error(err)
			return err
		}
		user, ok := m.sessionStorage.GetUser(session.UserAddr)
		if !ok {
			err := fmt.Errorf("user not found")
			sourceLog.Error(err)
			return err
		}
		userPubKey := user.PubKey
		isValid := m.morRpc.VerifySignature(msg.Params, signature, userPubKey, sourceLog)
		if !isValid {
			err := fmt.Errorf("invalid signature")
			sourceLog.Error(err)
			return err
		}

		var req *openai.ChatCompletionRequest

		err := json.Unmarshal([]byte(prompt), &req)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to unmarshal prompt"), err)
			sourceLog.Error(err)
			return err
		}

		_, err = m.apibus.PromptStream(ctx, req, func(response *openai.ChatCompletionStreamResponse) error {
			marshalledResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}

			encryptedResponse, err := lib.EncryptString(string(marshalledResponse), userPubKey)
			if err != nil {
				return err
			}

			// Send response
			r, err := m.morRpc.SessionPromptResponse(
				encryptedResponse,
				m.privateKeyHex,
				requestId,
			)
			if err != nil {
				err := lib.WrapError(fmt.Errorf("failed to create response"), err)
				sourceLog.Error(err)
				return err
			}
			sendCallback(r)
			return nil
		})

		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to prompt"), err)
			sourceLog.Error(err)
			return err
		}

		return nil
	default:
		err := fmt.Errorf("unknown method: %s", msg.Method)
		sourceLog.Error(err)
		return err
	}
}
