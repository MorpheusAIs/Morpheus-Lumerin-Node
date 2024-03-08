package httphandlers

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/contract"
	hrcontract "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/contract"
	"golang.org/x/exp/slices"
)

func (h *HTTPHandler) CreateContract(ctx *gin.Context) {
	dest, err := url.Parse(ctx.Query("dest"))
	if err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	hrGHS, err := strconv.ParseInt(ctx.Query("hrGHS"), 10, 0)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	duration, err := time.ParseDuration(ctx.Query("duration"))
	if err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	now := time.Now()
	destEnc, err := lib.EncryptString(dest.String(), h.pubKey)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	terms := hashrate.NewTerms(
		lib.GetRandomAddr().String(),
		lib.GetRandomAddr().String(),
		lib.GetRandomAddr().String(),
		now,
		duration,
		float64(hrGHS)*1e9,
		big.NewInt(0),
		0,
		hashrate.BlockchainStateRunning,
		false,
		big.NewInt(0),
		false,
		0,
		destEnc,
		"",
		"",
	)
	h.contractManager.AddContract(context.Background(), terms)

	ctx.JSON(200, gin.H{"status": "ok"})
}

func (c *HTTPHandler) GetContracts(ctx *gin.Context) {
	data := []Contract{}
	var errOuter error

	c.contractManager.GetContracts().Range(func(item resources.Contract) bool {
		contract, err := c.mapContract(ctx, item)
		if err != nil {
			errOuter = err
			return false
		}
		data = append(data, *contract)
		return true
	})

	if errOuter != nil {
		ctx.JSON(500, gin.H{"error": errOuter.Error()})
		return
	}

	slices.SortStableFunc(data, func(a Contract, b Contract) bool {
		return a.ID < b.ID
	})

	ctx.JSON(200, data)
}

type ContractsQP struct {
	IsDeleted         *bool   `form:"isDeleted"`
	HasFutureTerms    *bool   `form:"hasFutureTerms"`
	Role              *string `form:"role" validate:"omitempty,oneof=seller buyer"`
	BlockchainStatus  *string `form:"blockchainStatus" validate:"omitempty,oneof=available running"`
	ApplicationStatus *string `form:"applicationStatus" validate:"omitempty,oneof=pending running"`
	BuyerAddr         *string `form:"buyerAddr" validate:"omitempty,eth_addr"`
	SellerAddr        *string `form:"sellerAddr" validate:"omitempty,eth_addr"`
	ValidatorAddr     *string `form:"validatorAddr" validate:"omitempty,eth_addr"`
}

func (c *HTTPHandler) GetContractsV2(ctx *gin.Context) {
	qp := ContractsQP{}
	err := ctx.ShouldBindQuery(&qp)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = c.validator.StructCtx(ctx, qp)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	res := ContractsResponse{
		SellerTotal:    SellerTotal{},
		BuyerTotal:     BuyerTotal{},
		ValidatorTotal: BuyerTotal{},
		Contracts:      []Contract{},
	}

	var errOuter error

	c.contractManager.GetContracts().Range(func(item resources.Contract) bool {
		if qp.IsDeleted != nil && *qp.IsDeleted && !item.IsDeleted() {
			return true
		}

		if qp.IsDeleted != nil && !*qp.IsDeleted && item.IsDeleted() {
			return true
		}

		if qp.HasFutureTerms != nil && !*qp.HasFutureTerms && item.HasFutureTerms() {
			return true
		}

		if qp.HasFutureTerms != nil && *qp.HasFutureTerms && !item.HasFutureTerms() {
			return true
		}

		if qp.Role != nil && item.Role().String() != *qp.Role {
			return true
		}

		if qp.BlockchainStatus != nil && item.BlockchainState().String() != *qp.BlockchainStatus {
			return true
		}

		if qp.BuyerAddr != nil && item.Buyer() != *qp.BuyerAddr {
			return true
		}

		if qp.SellerAddr != nil && item.Seller() != *qp.SellerAddr {
			return true
		}

		if qp.ValidatorAddr != nil && item.Validator() != *qp.ValidatorAddr {
			return true
		}

		cnt, err := c.mapContract(ctx, item)
		if err != nil {
			errOuter = err
			return false
		}
		res.Contracts = append(res.Contracts, *cnt)

		if item.Role() == resources.ContractRoleSeller { // readonly
			res.SellerTotal.TotalNumber++
			res.SellerTotal.TotalBalanceLMR += cnt.BalanceLMR

			if item.BlockchainState() == hashrate.BlockchainStateRunning { // readonly
				res.SellerTotal.RunningNumber++
				res.SellerTotal.RunningTargetGHS += int(item.ResourceEstimates()[contract.ResourceEstimateHashrateGHS]) // readonly
				res.SellerTotal.RunningActualGHS += int(item.ResourceEstimatesActual()[c.hashrateCounterDefault])       // multiple atomics

				if item.StarvingGHS() > 0 { // atomic
					res.SellerTotal.StarvingNumber++
					res.SellerTotal.StarvingGHS += item.StarvingGHS()
				}
			}

			if item.BlockchainState() == hashrate.BlockchainStateAvailable { // readonly
				res.SellerTotal.AvailableNumber++
				res.SellerTotal.AvailableGHS += int(item.ResourceEstimates()[contract.ResourceEstimateHashrateGHS])
			}
		}

		if item.Role() == resources.ContractRoleBuyer { // readonly
			res.BuyerTotal.Number++
			res.BuyerTotal.HashrateGHS += int(item.ResourceEstimates()[contract.ResourceEstimateHashrateGHS])
			res.BuyerTotal.ActualHashrateGHS += int(item.ResourceEstimatesActual()[c.hashrateCounterDefault]) // readonly
			res.BuyerTotal.StarvingGHS += item.StarvingGHS()                                                  // atomic
		}

		if item.Role() == resources.ContractRoleValidator { // readonly
			res.ValidatorTotal.Number++
			res.ValidatorTotal.HashrateGHS += int(item.ResourceEstimates()[contract.ResourceEstimateHashrateGHS])
			res.ValidatorTotal.ActualHashrateGHS += int(item.ResourceEstimatesActual()[c.hashrateCounterDefault]) // readonly
		}

		return true
	})

	if errOuter != nil {
		ctx.JSON(500, gin.H{"error": errOuter})
		return
	}

	slices.SortStableFunc(res.Contracts, func(a Contract, b Contract) bool {
		return a.ID < b.ID
	})

	ctx.JSON(200, res)
}

func (c *HTTPHandler) GetContract(ctx *gin.Context) {
	contractID := ctx.Param("ID")
	if contractID == "" {
		ctx.JSON(400, gin.H{"error": "contract id is required"})
		return
	}
	contract, ok := c.contractManager.GetContracts().Load(contractID)
	if !ok {
		ctx.JSON(404, gin.H{"error": "contract not found"})
		return
	}

	contractData, err := c.mapContract(ctx, contract)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, contractData)
}

func (c *HTTPHandler) GetDeliveryLogsConsole(ctx *gin.Context) {
	contractID := ctx.Param("ID")
	if contractID == "" {
		ctx.JSON(400, gin.H{"error": "contract id is required"})
		return
	}
	contract, ok := c.logStorage.Load(contractID)
	if !ok {
		ctx.JSON(404, gin.H{"error": "contract not found"})
		return
	}

	ctx.Status(200)
	_, err := io.Copy(ctx.Writer, contract.GetReader())
	if err != nil {
		c.log.Errorf("failed to write logs: %s", err)
		_ = ctx.Error(err)
		ctx.Abort()
	}
}

func (c *HTTPHandler) GetDeliveryLogs(ctx *gin.Context) {
	contractID := ctx.Param("ID")
	if contractID == "" {
		ctx.JSON(400, gin.H{"error": "contract id is required"})
		return
	}
	contract, ok := c.contractManager.GetContracts().Load(contractID)
	if !ok {
		ctx.JSON(404, gin.H{"error": "contract not found"})
		return
	}

	sellerContract, ok := contract.(*hrcontract.ControllerSeller)
	if !ok {
		ctx.JSON(400, gin.H{"error": "contract is not seller contract"})
		return
	}
	logs, err := sellerContract.GetDeliveryLogs()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = writeHTML(ctx.Writer, logs)
	if err != nil {
		c.log.Errorf("failed to write logs: %s", err)
		_ = ctx.Error(err)
		ctx.Abort()
	}
	return
}

func (p *HTTPHandler) mapContract(ctx context.Context, item resources.Contract) (*Contract, error) {

	return &Contract{
		Resource: Resource{
			Self: p.publicUrl.JoinPath(fmt.Sprintf("/contracts/%s", item.ID())).String(), // readonly
		},
		Logs:        p.publicUrl.JoinPath(fmt.Sprintf("/contracts/%s/logs", item.ID())).String(),         // readonly
		ConsoleLogs: p.publicUrl.JoinPath(fmt.Sprintf("/contracts/%s/logs-console", item.ID())).String(), // readonly

		Role:                    item.Role().String(),                                   // readonly
		Stage:                   item.ValidationStage().String(),                        // atomic
		ID:                      item.ID(),                                              // readonly
		BuyerAddr:               item.Buyer(),                                           // readonly
		ValidatorAddr:           item.Validator(),                                       // readonly
		SellerAddr:              item.Seller(),                                          // readonly
		ResourceEstimatesTarget: roundResourceEstimates(item.ResourceEstimates()),       // readonly
		ResourceEstimatesActual: roundResourceEstimates(item.ResourceEstimatesActual()), // multiple atomics
		StarvingGHS:             item.StarvingGHS(),                                     // atomic
		PriceLMR:                LMRWithDecimalsToLMR(item.Price()),                     // readonly
		ProfitTarget:            item.ProfitTarget(),                                    // readonly
		Duration:                formatDuration(item.Duration()),                        // readonly

		IsDeleted:      item.IsDeleted(),                     // readonly
		BalanceLMR:     LMRWithDecimalsToLMR(item.Balance()), // readonly
		HasFutureTerms: item.HasFutureTerms(),                // readonly
		Version:        item.Version(),                       // readonly

		StartTimestamp:    formatTime(item.StartTime()),    // readonly
		EndTimestamp:      formatTime(item.EndTime()),      // readonly
		Elapsed:           formatDuration(item.Elapsed()),  // readonly
		ApplicationStatus: item.State().String(),           // rw mutex canceable
		BlockchainStatus:  item.BlockchainState().String(), // readonly
		Error:             errString(item.Error()),         // atomic
		Dest:              item.Dest(),                     // readonly
		PoolDest:          item.PoolDest(),                 // readonly
		// Miners:            p.allocator.GetMinersFulfillingContract(item.ID(), p.cycleDuration),
	}, nil
}

func errString(s error) string {
	if s != nil {
		return s.Error()
	}
	return ""
}

func writeHTML(w io.Writer, logs []hrcontract.DeliveryLogEntry) error {
	header := []string{
		"TimestampUnix",
		"ActualGHS",
		"FullMinersGHS",
		"FullMiners",
		"FullMinersShares",
		"PartialMinersGHS",
		"PartialMiners",
		"PartialMinersShares",
		"UnderDeliveryGHS",
		"GlobalHashrateGHS",
		"GlobalUnderDeliveryGHS",
		"GlobalError",
		"NextCyclePartialDeliveryTargetGHS",
	}

	// header
	_, _ = w.Write([]byte(`
		<html>
			<style>
				table {
					font-family: monospace;
					border-collapse: collapse;
					font-size: 12px;
					border: 1px solid #333;
				}
				th, td {
					padding: 3px;
					border: 1px solid #333;
				}
			</style>
			<body>
				<table>`))

	// table header
	_, _ = w.Write([]byte("<tr>"))
	for _, h := range header {
		err := writeTableRow("th", w, h)
		if err != nil {
			return err
		}
	}
	_, _ = w.Write([]byte("</tr>"))

	// table body
	for _, entry := range logs {
		_, _ = w.Write([]byte("<tr>"))
		err := writeTableRow("td", w,
			formatTime(entry.Timestamp),
			fmt.Sprint(entry.ActualGHS),
			fmt.Sprint(entry.FullMinersGHS),
			fmt.Sprint(entry.FullMiners),
			fmt.Sprint(entry.FullMinersShares),
			fmt.Sprint(entry.PartialMinersGHS),
			fmt.Sprint(entry.PartialMiners),
			fmt.Sprint(entry.PartialMinersShares),
			fmt.Sprint(entry.UnderDeliveryGHS),
			fmt.Sprint(entry.GlobalHashrateGHS),
			fmt.Sprint(entry.GlobalUnderDeliveryGHS),
			fmt.Sprintf("%.2f", entry.GlobalError),
			fmt.Sprint(entry.NextCyclePartialDeliveryTargetGHS),
		)
		if err != nil {
			return err
		}
		_, _ = w.Write([]byte("</tr>"))
	}

	// footer
	_, _ = w.Write([]byte(`
				</table>
			</body>
		</html>`))

	return nil
}

func writeTableRow(tag string, w io.Writer, values ...string) error {
	for _, value := range values {
		_, err := w.Write([]byte(fmt.Sprintf("<%s>%s</%s>", tag, value, tag)))
		if err != nil {
			return err
		}
	}
	return nil
}
