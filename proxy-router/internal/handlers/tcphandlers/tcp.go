package tcphandlers

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	morrpc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/transport"
)

func NewTCPHandler(
	tcpLog lib.ILogger,
	morRpcHandler *proxyapi.MORRPCController,
) transport.Handler {
	return func(ctx context.Context, conn net.Conn) {
		if tcp, ok := conn.(*net.TCPConn); ok {
			err := tcp.SetKeepAlive(true)
			if err != nil {
				tcpLog.Errorf("Error setting keepalive: %s", err)
			}
			err = tcp.SetKeepAlivePeriod(10 * time.Second)
			if err != nil {
				tcpLog.Errorf("Error setting keepalive period: %s", err)
			}
		}

		addr := conn.RemoteAddr().String()
		sourceLog := tcpLog.Named("TCP").With("SrcAddr", addr)

		defer func() {
			sourceLog.Debugf("closing connection")
			conn.Close()
		}()

		msg, err := getMessage(conn)
		if err != nil {
			sourceLog.Error("error reading message", err)
			return
		}

		err = morRpcHandler.Handle(ctx, *msg, sourceLog, func(resp *morrpc.RpcResponse) error {
			sourceLog.Debugf("sending TCP response for method: %s", msg.Method)
			_, err := sendMsg(conn, resp)
			if err != nil {
				sourceLog.Errorf("Error sending message: %s", err)
				return err
			}
			return nil
		})
		if err != nil {
			sourceLog.Errorf("Error handling message with method: %s: %s", msg.Method, err)
			// Send unencrypted error response for infrastructure errors
			// Empty/nil signature indicates this error is not encrypted
			errorResp := &morrpc.RpcResponse{
				ID: msg.ID,
				Error: &morrpc.RpcError{
					Code:    400,
					Message: err.Error(),
					Data: morrpc.RPCErrorData{
						Timestamp: uint64(time.Now().UnixMilli()),
						Signature: lib.MustStringToHexString("0x0"),
					},
				},
			}
			sendMsg(conn, errorResp)
			return
		}
	}
}

func sendMsg(conn net.Conn, msg *morrpc.RpcResponse) (int, error) {
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	return conn.Write(msgJson)
}

func getMessage(conn net.Conn) (*morrpc.RPCMessage, error) {
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)

	var msg *morrpc.RPCMessage
	err := d.Decode(&msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
