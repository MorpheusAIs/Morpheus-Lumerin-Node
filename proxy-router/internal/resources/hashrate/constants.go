package hashrate

type ValidationStage int8

const (
	ValidationStageNotApplicable ValidationStage = 0
	ValidationStageNotValidating ValidationStage = 1
	ValidationStageValidating    ValidationStage = 2
	ValidationStageFinished      ValidationStage = 3
)

func (s ValidationStage) String() string {
	switch s {
	case ValidationStageNotValidating:
		return "not validating"
	case ValidationStageValidating:
		return "validating"
	case ValidationStageFinished:
		return "finished"
	case ValidationStageNotApplicable:
		return "not applicable"
	default:
		return "unknown"
	}
}

type BlockchainState int

const (
	BlockchainStateAvailable BlockchainState = 0
	BlockchainStateRunning   BlockchainState = 1
)

func (b BlockchainState) String() string {
	switch b {
	case BlockchainStateAvailable:
		return "available"
	case BlockchainStateRunning:
		return "running"
	default:
		return "unknown"
	}
}
