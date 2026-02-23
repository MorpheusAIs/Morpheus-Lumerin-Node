package rating

import (
	"math"

	s "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
)

// ratioScore calculates the ratio of two numbers
func ratioScore(num, denom uint32) float64 {
	if denom == 0 {
		return 0
	}
	return float64(num) / float64(denom)
}

// normZIndex normalizes the value using z-index
func normZIndex(pmMean int64, mSD s.LibSDSD, obsNum int64) float64 {
	sd := getSD(mSD, obsNum)
	if sd == 0 {
		return 0
	}
	// TODO: consider variance(SD) of provider model stats
	return float64(pmMean-mSD.Mean) / sd
}

// normRange normalizes the incoming data within the range [-normRange, normRange]
// cutting off the values outside the range, and then shifts and scales to the range [0, 1]
func normRange(input float64, normRange float64) float64 {
	return cutRange01((input + normRange) / (2.0 * normRange))
}

// getSD calculates the standard deviation from the standard deviation struct
func getSD(sd s.LibSDSD, obsNum int64) float64 {
	variance := getVariance(sd, obsNum)
	if variance <= 0 {
		return 0
	}
	return math.Sqrt(variance)
}

// getVariance calculates the variance from the standard deviation struct
func getVariance(sd s.LibSDSD, obsNum int64) float64 {
	if obsNum <= 1 {
		return 0
	}
	return float64(sd.SqSum) / float64(obsNum-1)
}

// cutRange01 cuts the value to the range [0, 1]
func cutRange01(val float64) float64 {
	if val > 1 {
		return 1
	}
	if val < 0 {
		return 0
	}
	return val
}

// normMinMax normalizes the value within the range [min, max] to the range [0, 1]
func normMinMax(val, min, max int64) float64 {
	if max == min {
		return 0
	}
	if val < min {
		return 0
	}
	if val > max {
		return 1
	}
	return float64(val-min) / float64(max-min)
}
