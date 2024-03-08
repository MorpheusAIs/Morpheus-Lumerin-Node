package system

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestGetFileDescriptors(t *testing.T) {
	fds, err := NewOSConfigurator().GetFileDescriptors(context.TODO(), os.Getpid())
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("fds: %+v\n", fds)
}
