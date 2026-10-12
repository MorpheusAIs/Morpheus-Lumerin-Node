package proxyapi

func sessionExpiredByServerTime(endsAtUnixSec uint64, nowMs uint64) bool {
	return endsAtUnixSec*1000 < nowMs
}
