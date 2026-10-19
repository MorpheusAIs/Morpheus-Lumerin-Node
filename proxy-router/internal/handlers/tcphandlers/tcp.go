package tcphandlers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"syscall"
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
			logInboundReadErr(sourceLog, err)
			return
		}

		sourceLog = sourceLog.With("request_id", msg.ID)

		err = morRpcHandler.Handle(ctx, *msg, sourceLog, func(resp *morrpc.RpcResponse) error {
			sourceLog.Infof("sending TCP response for method: %s", msg.Method)
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

// logInboundReadErr chooses a level for a connection that never produced a
// MOR-RPC message. Error attaches a stack trace, and scanners hit this path
// constantly, so only an unexpected failure uses it.
func logInboundReadErr(log lib.ILogger, err error) {
	switch {
	case peerClosed(err):
		log.Debugf("connection closed before a message: %s", err)
	case malformedMessage(err):
		log.Warnf("ignored malformed message: %s", err)
	default:
		log.Error("error reading message", err)
	}
}

// peerClosed reports a peer that went away without a usable frame.
func peerClosed(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE)
}

// malformedMessage reports bytes that are not a MOR-RPC message. Port probes
// look like this; a stack trace does not make them easier to read.
func malformedMessage(err error) bool {
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var syntax *json.SyntaxError
	var wrongType *json.UnmarshalTypeError
	return errors.As(err, &syntax) || errors.As(err, &wrongType)
}
