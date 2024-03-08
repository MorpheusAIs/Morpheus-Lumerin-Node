package lib

import (
	"net"
)

func TCPPipe() (clientConn net.Conn, serverConn net.Conn, err error) {
	server, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, nil, err
	}

	errCh := make(chan error, 1)

	go func() {
		serverConn, err = server.Accept()
		if err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	clientConn, err = net.DialTCP("tcp", nil, server.Addr().(*net.TCPAddr))
	if err != nil {
		return nil, nil, err
	}

	return clientConn, serverConn, <-errCh
}

func ParsePort(addr string) (port string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return port
}
