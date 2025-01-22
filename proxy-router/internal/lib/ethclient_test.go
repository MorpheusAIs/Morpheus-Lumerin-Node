package lib

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestDeco(t *testing.T) {
	inp := "0x095ea7b3000000000000000000000000b8c55cd613af947e73e262f0d3c54b7211af16cf000000000000000000000000000000000000000000000000000000003b9aca00"
	inpHex := common.FromHex(inp)
	methodName, entries, err := DecodeInput(inpHex)
	require.NoError(t, err)
	fmt.Println(methodName)
	fmt.Println(entries)
}
