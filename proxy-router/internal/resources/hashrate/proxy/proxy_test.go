package proxy

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
	m "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/validator"
)

func RunTestProxy() (p *Proxy, s *StratumConnection, d *StratumConnection, cancel func(), done chan error) {
	sourceClient, source := net.Pipe()
	destClient, dest := net.Pipe()
	timeout := 10 * time.Minute

	log := lib.NewTestLogger()
	noLog := &lib.LoggerMock{}
	destURL, _ := url.Parse("stratum+tcp://test:test@localhost:3333")
	log.Warnf("started server")

	sourceConn := NewSourceConn(CreateConnection(sourceClient, "", timeout, timeout, log), log)
	valid := validator.NewValidator(time.Minute)
	destConn := NewDestConn(CreateConnection(destClient, destURL.String(), timeout, timeout, log), valid, destURL, log)

	destConnFactory := func(ctx context.Context, url *url.URL, srcWorker string, srcAddr string) (*ConnDest, error) {
		return destConn, nil
	}
	hashrateFactory := func() *hashrate.Hashrate {
		return hashrate.NewHashrate(map[string]hashrate.Counter{})
	}

	globalHashrate := hashrate.NewGlobalHashrate(hashrateFactory)

	proxy := NewProxy("test", sourceConn, destConnFactory, hashrateFactory, globalHashrate, destURL, true, 1, 5, log, func(id string) (resources.Contract, bool) {
		return nil, false
	})

	ctx, cancel := context.WithCancel(context.Background())
	runErrorCh := make(chan error)

	go func() {
		defer sourceClient.Close()
		defer destClient.Close()
		defer close(runErrorCh)

		err := proxy.Connect(ctx)
		if err != nil {
			log.Errorf("proxy exited with error %v", err)
			runErrorCh <- err
			return
		}

		err = proxy.Run(ctx)
		if !errors.Is(err, context.Canceled) {
			log.Errorf("proxy exited with error %v", err)
		} else {
			log.Info("proxy exited")
		}
		runErrorCh <- err
	}()

	return proxy,
		CreateConnection(source, "", timeout, timeout, noLog),
		CreateConnection(dest, "", timeout, timeout, noLog), cancel, runErrorCh
}

func TestHandshakeStartWithMiningConfigure(t *testing.T) {
	_, src, dest, cancel, errCh := RunTestProxy()
	defer func() {
		cancel()
		<-errCh
	}()

	// mining.configure
	err := src.Write(context.TODO(), m.NewMiningConfigure(1, &m.MiningConfigureExtensionParams{
		VersionRollingMask:        "00000000",
		VersionRollingMinBitCount: 2,
	}))
	require.NoError(t, err)

	msg, err := dest.Read(context.Background())
	require.NoError(t, err)

	require.IsType(t, &m.MiningConfigure{}, msg)
	mask, bits := msg.(*m.MiningConfigure).GetVersionRolling()
	require.Equal(t, "00000000", mask)
	require.Equal(t, 2, bits)

	// mining.configure result
	err = dest.Write(context.Background(), m.NewMiningConfigureResult(1, true, "00000000"))
	require.NoError(t, err)

	msg, err = src.Read(context.Background())
	require.NoError(t, err)
	require.IsType(t, &m.MiningResult{}, msg)
}

func TestHandshakeStartWithMiningSubscribe(t *testing.T) {
	_, src, dest, cancel, errCh := RunTestProxy()
	defer func() {
		cancel()
		<-errCh
	}()

	// mining.subscribe
	err := src.Write(context.TODO(), m.NewMiningSubscribe(1, "test", "test"))
	require.NoError(t, err)

	msg, err := dest.Read(context.Background())
	require.NoError(t, err)

	require.IsType(t, &m.MiningSubscribe{}, msg)
	require.Equal(t, 1, msg.(*m.MiningSubscribe).GetID())
	require.Equal(t, "test", msg.(*m.MiningSubscribe).GetUseragent())
	require.Equal(t, "test", msg.(*m.MiningSubscribe).GetWorkerNumber())

	// mining.subscribe result
	err = dest.Write(context.Background(), m.NewMiningSubscribeResult(1, "11650803bc7550", 8))
	require.NoError(t, err)

	msg, err = src.Read(context.Background())
	require.NoError(t, err)
	require.IsType(t, &m.MiningResult{}, msg)

	typed, err := m.ToMiningSubscribeResult(msg.(*m.MiningResult))
	require.NoError(t, err)

	xn, size := typed.GetExtranonce()
	require.Equal(t, "11650803bc7550", xn)
	require.Equal(t, 8, size)
}

func TestSourceMessageInvalid(t *testing.T) {}

func TestDestinationMessageInvalid(t *testing.T) {}

func TestDestinationReplyTimeout(t *testing.T) {}

func TestDestinationReplyError(t *testing.T) {}

func TestDestinationReadTimeout(t *testing.T) {}

func TestDestinationWriteTimeout(t *testing.T) {}

func TestSourceReadTimeout(t *testing.T) {}

func TestSourceWriteTimeout(t *testing.T) {}

func TestSourceClose(t *testing.T) {}

func TestDestClose(t *testing.T) {}

func TestDestChange(t *testing.T) {}

func TestDestChangeFailure(t *testing.T) {}

func TestDestChangeTimeout(t *testing.T) {}

func TestDestChangeInvalidDest(t *testing.T) {}

func TestDestChangeWithMiningConfigure(t *testing.T) {}

func TestDestChangeWithoutMiningConfigure(t *testing.T) {}

func TestDestChangeVersionMaskNegotiation(t *testing.T) {}

func TestDestChangeVersionMaskNegotiationFailure(t *testing.T) {}

func TestDestChangeResetJob(t *testing.T) {}

func TestDestConnectionCachedClosureOnTimeout(t *testing.T) {}

func TestDestConnectionCachedReuse(t *testing.T) {}

func TestDestConnectionCachedKeepReading(t *testing.T) {}

func TestHashrateCount(t *testing.T) {}

func TestInvalidSubmit(t *testing.T) {}

func TestValidSubmitLowDiff(t *testing.T) {}

func TestSetDifficultyAppliesOnlyToNewJobs(t *testing.T) {
	// test that new difficulty is applied to jobs that has been received after the last set_difficulty message
}

func TestVerifyJobWithCorrespondingDifficulty(t *testing.T) {
	// check that validator uses difficulty that was current at the moment of receiving the job
}

func TestHandshakeMessagesComeInCorrectOrder(t *testing.T) {
	// check that mining.set_version_mask comes after mining.configure result
	// check that mining.set_extranonce comes after mining.subscribe result
}
