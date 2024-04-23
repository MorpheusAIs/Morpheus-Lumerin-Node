package tcphandlers

import (
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/morrpc"
)

type MorRpcHandler struct {
	privateKeyHex string
	publicKeyHex  string
	address       string
	morRpc        *morrpc.MorRpc
}

func NewMorRpcHandler(privateKeyHex string, publicKeyHex string, address string, morRpc *morrpc.MorRpc) *MorRpcHandler {
	return &MorRpcHandler{
		privateKeyHex: privateKeyHex,
		address:       address,
		publicKeyHex:  publicKeyHex,
		morRpc:        morRpc,
	}
}

func (m *MorRpcHandler) Handle(msg morrpc.RpcMessage, sourceLog interfaces.ILogger) (*morrpc.RpcResponse, error) {
	switch msg.Method {
	case "session.request":
		requestId := fmt.Sprintf("%v", msg.ID)
		signature := fmt.Sprintf("%v", msg.Params["signature"])
		userAddr := fmt.Sprintf("%v", msg.Params["user"])
		userPubKey := fmt.Sprintf("%v", msg.Params["key"])
		spend := fmt.Sprintf("%v", msg.Params["spend"])
		timeStamp := fmt.Sprintf("%v", msg.Params["timestamp"])
		sourceLog.Debugf("Received session request from %s, timestamp: %s", userAddr, timeStamp)

		isValid := m.morRpc.VerifySignature(msg.Params, signature, userPubKey, sourceLog)
		if !isValid {
			err := fmt.Errorf("invalid signature")
			sourceLog.Error(err)
			return nil, err
		}

		hasCapacity := true // check if there is capacity
		if !hasCapacity && spend != "" {
			err := fmt.Errorf("no capacity")
			sourceLog.Error(err)
			return nil, err
		}

		// Send response
		response, err := m.morRpc.InitiateSessionResponse(
			m.publicKeyHex,
			userAddr,
			m.privateKeyHex,
			requestId,
		)
		if err != nil {
			err := lib.WrapError(fmt.Errorf("failed to create response"), err)
			sourceLog.Error(err)
			return nil, err
		}

		return response, nil
	case "method2":
		// handle method2
	default:
		err := fmt.Errorf("unknown method: %s", msg.Method)
		sourceLog.Error(err)
		return nil, err
	}
	return nil, nil
}
