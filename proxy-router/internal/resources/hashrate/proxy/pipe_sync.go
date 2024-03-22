package proxy

import (
	"context"

	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
)

type pipeSync struct {
	stream1            i.StratumReadWriter
	stream2            i.StratumReadWriter
	interceptor1       Interceptor
	interceptor2       Interceptor
	startStream2Signal chan struct{}
}

// Provides a way to synchronize two i.StratumReadWriters. Only one message
// from either stream can be processed at a time, it gives deternimistic
// message write order
func NewPipeSync(stream1, stream2 i.StratumReadWriter, interceptor1, interceptor2 Interceptor) *pipeSync {
	pipe := &pipeSync{
		stream1:            stream1,
		stream2:            stream2,
		interceptor1:       interceptor1,
		interceptor2:       interceptor2,
		startStream2Signal: make(chan struct{}),
	}

	return pipe
}

func (p *pipeSync) Run(ctx context.Context) error {
	return pipeDuplexSync(ctx, p.getStream1, p.getStream2, p.interceptor1, p.interceptor2, p.startStream2Signal)
}

func (p *pipeSync) SetStream2(readWriter i.StratumReadWriter) {
	p.stream2 = readWriter
}

func (p *pipeSync) StartStream2() {
	close(p.startStream2Signal)
}

func pipeDuplexSync(ctx context.Context, stream1, stream2 func() i.StratumReadWriter, interceptor1, interceptor2 Interceptor, startStream2Signal chan struct{}) error {
	stream1Ch, stream2Ch := make(chan i.MiningMessageGeneric), make(chan i.MiningMessageGeneric)
	stream1Err, stream2Err := make(chan error, 1), make(chan error, 1)
	stream1Ctx, stream1Cancel := context.WithCancel(ctx)
	stream2Ctx, stream2Cancel := context.WithCancel(ctx)

	go func() {
		for {
			err := readToChan(stream1Ctx, stream1, stream1Ch)
			if err != nil {
				stream1Err <- err
				close(stream1Err)
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-startStream2Signal:
			case <-stream2Ctx.Done():
				stream2Err <- stream2Ctx.Err()
				close(stream2Err)
				return
			}

			err := readToChan(stream2Ctx, stream2, stream2Ch)
			if err != nil {
				stream2Err <- err
				close(stream2Err)
				return
			}
		}
	}()

	cancel := func() {
		stream2Cancel()
		stream1Cancel()
		<-stream2Err
		<-stream1Err
	}

	// in case of any error, cancel all streams
	defer cancel()

	for {
		var msg i.MiningMessageGeneric
		select {
		case msg = <-stream1Ch:
			outMsg, err := interceptor1(ctx, msg)
			if err != nil {
				return err
			}
			if outMsg != nil {
				err := stream2().Write(ctx, outMsg)
				if err != nil {
					return err
				}
			}
		case msg = <-stream2Ch:
			outMsg, err := interceptor2(ctx, msg)
			if err != nil {
				return err
			}
			if outMsg != nil {
				err := stream1().Write(ctx, outMsg)
				if err != nil {
					return err
				}
			}
			// todo: write if msg is not nil
		case err := <-stream1Err:
			return err
		case err := <-stream2Err:
			return err
		}
	}
}

func (p *pipeSync) getStream1() i.StratumReadWriter {
	return p.stream1
}

func (p *pipeSync) getStream2() i.StratumReadWriter {
	return p.stream2
}

func readToChan(ctx context.Context, stream func() i.StratumReadWriter, ch chan<- i.MiningMessageGeneric) error {
	for {
		msg, err := stream().Read(ctx)
		if err != nil {
			return err
		}
		// nil check is required because message could be intercepted on connection level
		// it returns nil to tell that nothing should be written
		if msg != nil {
			select {
			case ch <- msg:
			case <-ctx.Done():
			}
		}
	}
}
