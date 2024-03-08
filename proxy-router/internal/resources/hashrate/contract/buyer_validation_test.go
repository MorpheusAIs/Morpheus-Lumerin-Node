package contract

import (
	"fmt"
	"testing"
	"time"
)

func TestC1(t *testing.T) {
	data := []time.Duration{}
	for i := 0; i < 60; i++ {
		data = append(data, time.Duration(i)*time.Minute*10)
	}

	for _, d := range data {
		k := GetMaxGlobalError(d, 0.05, 20*time.Minute, 5*time.Minute)
		fmt.Printf("elapsed %s - error threshold %.2f\n", d, k)
	}
}
