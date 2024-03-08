package interfaces

import "time"

type Hashrate interface {
	GetTotalWork() float64
	GetTotalDuration() time.Duration
	GetLastSubmitTime() time.Time
	GetHashrate5minGHS() float64
	GetHashrateTotalGHS() float64
	GetHashrateCustomGHS(duration time.Duration) (float64, bool)
	GetDurationHashrateGHSMap() map[time.Duration]float64
}
