package proxy

// import (
// 	"context"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/require"
// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
// 	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
// )

// func TestPipeAsync(t *testing.T) {
// 	// Create a new pipe
// 	sourceServer, sourceClient, err := lib.TCPPipe()
// 	require.NoError(t, err)
// 	defer sourceServer.Close()
// 	defer sourceClient.Close()

// 	destServer, destClient, err := lib.TCPPipe()
// 	require.NoError(t, err)
// 	defer destServer.Close()
// 	defer destClient.Close()

// 	sourceClientConn := CreateConnection(sourceClient, "source-client", time.Minute, time.Minute, &lib.LoggerMock{})
// 	sourceServerConn := CreateConnection(sourceServer, "source-server", time.Minute, time.Minute, &lib.LoggerMock{})
// 	destClientConn := CreateConnection(destClient, "dest-client", time.Minute, time.Minute, &lib.LoggerMock{})
// 	destServerConn := CreateConnection(destServer, "dest-server", time.Minute, time.Minute, &lib.LoggerMock{})
// 	noopInterceptor := func(ctx context.Context, msg i.MiningMessageGeneric) (i.MiningMessageGeneric, error) { return msg, nil }
// 	pipe := NewPipe(sourceServerConn, destClientConn, noopInterceptor, noopInterceptor, &lib.LoggerMock{})

// 	ctx := context.Background()

// 	// start writing to the source
// 	go func() {
// 		for i := 0; i < 5; i++ {
// 			select {
// 			case <-time.After(7 * time.Second):
// 			case <-ctx.Done():
// 				fmt.Println("source write cancelled")
// 				return
// 			}
// 			err := sourceClientConn.Write(ctx, stratumv1_message.NewMiningSubscribe(0, "name", "0"))
// 			require.NoError(t, err)
// 			fmt.Println("source write")

// 		}
// 	}()

// 	// start reading the dest
// 	go func() {
// 		for {
// 			_, err := destServerConn.Read(ctx)
// 			require.NoError(t, err)
// 			fmt.Println("dest read")
// 		}
// 	}()

// 	// Create a new context
// 	pipe.StartSourceToDest(ctx)
// 	pipe.StartSourceToDest(ctx)

// 	childCtx, childCancel := context.WithCancel(ctx)
// 	go func() {
// 		err := pipe.Run(childCtx)
// 		fmt.Printf("pipe error: %v\n", err)
// 		return
// 	}()

// 	// Stop the pipe
// 	<-time.After(3 * time.Second)
// 	fmt.Println("stopping pipe")
// 	childCancel()
// 	fmt.Println("stopped, waiting 10s")

// 	<-time.After(10 * time.Second)
// }
