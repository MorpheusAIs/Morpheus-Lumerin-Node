package contract

import "time"

func GetMaxGlobalError(elapsed time.Duration, minError float64, flatness, skipPeriod time.Duration) float64 {
	maxErr := float64(flatness) / float64(elapsed+flatness-skipPeriod)
	if maxErr > 1 {
		return 1
	}
	if maxErr < minError {
		return minError
	}
	return maxErr
}
