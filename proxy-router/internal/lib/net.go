package lib

import (
	"net"
)

// TCPPipe returns a connected pair of TCP connections over the loopback
// interface. The caller owns both and is responsible for closing them.
//
// The accept goroutine deliberately writes to its own local variables and
// hands the result back over a channel. An earlier version assigned straight
// to the named return values, so the goroutine's write to `err` raced the
// parent's write to the same variable from net.DialTCP; `go test -race`
// fails the package on it. The channel receive orders the two writes.
func TCPPipe() (net.Conn, net.Conn, error) {
	server, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, nil, err
	}
	// The listener has served its purpose once the single connection is
	// accepted; the previous version leaked it for the life of the process.
	defer server.Close()

	type accepted struct {
		conn net.Conn
		err  error
	}
	acceptCh := make(chan accepted, 1)

	go func() {
		conn, err := server.Accept()
		acceptCh <- accepted{conn: conn, err: err}
	}()

	clientConn, err := net.DialTCP("tcp", nil, server.Addr().(*net.TCPAddr))
	if err != nil {
		return nil, nil, err
	}

	res := <-acceptCh
	if res.err != nil {
		clientConn.Close()
		return nil, nil, res.err
	}

	return clientConn, res.conn, nil
}

func ParsePort(addr string) (port string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return port
}
