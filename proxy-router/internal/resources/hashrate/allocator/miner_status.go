package allocator

type MinerStatus uint8

const (
	MinerStatusVetting       MinerStatus = iota // vetting period
	MinerStatusFree                             // serving default pool
	MinerStatusBusy                             // fully or partially serving contract(s)
	MinerStatusPartialBusy                      // partially serving contract(s)
	MinerStatusDisconnecting                    // error or connection closeout caused the miner to disconnect, it might be briefly available in miners collection
)

func (m MinerStatus) String() string {
	switch m {
	case MinerStatusVetting:
		return "vetting"
	case MinerStatusFree:
		return "free"
	case MinerStatusBusy:
		return "busy"
	case MinerStatusPartialBusy:
		return "partial_busy"
	case MinerStatusDisconnecting:
		return "disconnecting"
	}
	// shouldn't reach here
	return "ERROR"
}
