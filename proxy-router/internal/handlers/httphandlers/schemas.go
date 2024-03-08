package httphandlers

import (
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/allocator"
)

type MinersResponse struct {
	TotalHashrateGHS     int
	UsedHashrateGHS      int
	AvailableHashrateGHS int

	TotalMiners       int
	VettingMiners     int
	FreeMiners        int
	PartialBusyMiners int
	BusyMiners        int

	Miners []Miner
}

type ContractsResponse struct {
	SellerTotal    SellerTotal
	BuyerTotal     BuyerTotal
	ValidatorTotal BuyerTotal

	Contracts []Contract
}

type ConfigResponse struct {
	Version       string
	Commit        string
	DerivedConfig interface{}
	Config        interface{}
}

type SellerTotal struct {
	TotalNumber     int
	TotalGHS        int
	TotalBalanceLMR float64

	RunningNumber    int
	RunningTargetGHS int
	RunningActualGHS int

	StarvingNumber int
	StarvingGHS    int

	AvailableNumber int
	AvailableGHS    int
}

type BuyerTotal struct {
	Number            int
	HashrateGHS       int
	ActualHashrateGHS int
	StarvingGHS       int
}

type Miner struct {
	Resource

	ID                    string
	WorkerName            string
	Status                string
	HashrateAvgGHS        map[string]int
	CurrentDestination    string
	CurrentDifficulty     int
	ConnectedAt           string
	Uptime                string
	ActivePoolConnections *map[string]string `json:",omitempty"`
	Destinations          []*allocator.DestItem
	Stats                 interface{}
}

type Contract struct {
	Resource

	Logs                    string
	ConsoleLogs             string
	Role                    string
	Stage                   string
	ID                      string
	BuyerAddr               string
	SellerAddr              string
	ValidatorAddr           string
	ResourceEstimatesTarget map[string]int
	ResourceEstimatesActual map[string]int
	StarvingGHS             int

	BalanceLMR     float64
	IsDeleted      bool
	HasFutureTerms bool
	Version        uint32

	StartTimestamp string
	EndTimestamp   string
	Duration       string
	PriceLMR       float64
	ProfitTarget   int8

	Elapsed           string
	ApplicationStatus string
	BlockchainStatus  string
	Error             string
	Dest              string
	PoolDest          string
	Miners            []*allocator.MinerItemJobScheduled
}

type Resource struct {
	Self string
}

type Worker struct {
	WorkerName string
	Hashrate   map[string]float64
	Reconnects int
}
