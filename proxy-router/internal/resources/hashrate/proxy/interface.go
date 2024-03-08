package proxy

import (
	"context"
	"net/url"
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
	i "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/interfaces"
	m "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/proxy/stratumv1_message"
)

type HashrateCounter interface {
	OnSubmit(diff float64)
}

type HashrateCounterFunc func(diff float64)

type GlobalHashrateCounter interface {
	OnSubmit(workerName string, diff float64)
	OnConnect(workerName string)
}

type Hashrate interface {
	GetHashrateAvgGHSCustom(ID string) (hrGHS float64, ok bool)
	GetHashrateAvgGHSAll() map[string]float64
	GetTotalWork() float64
	GetTotalDuration() time.Duration
	GetLastSubmitTime() time.Time
	GetTotalShares() int
}

type DestConnFactory = func(ctx context.Context, url *url.URL, srcWorker, srcAddr string) (*ConnDest, error)

type Interceptor = func(context.Context, i.MiningMessageGeneric) (i.MiningMessageGeneric, error)

type ResultHandler = func(a *m.MiningResult) (msg i.MiningMessageWithID, err error)

type HashrateFactory = func() *hashrate.Hashrate

type GetContractFromStoreFn func(id string) (resources.Contract, bool)
