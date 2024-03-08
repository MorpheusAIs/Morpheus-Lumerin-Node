package hashrate

import (
	"fmt"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	job := 9535809.0
	hs := JobSubmittedToGHSV2(job, 5*time.Minute)
	fmt.Println(hs)
}
