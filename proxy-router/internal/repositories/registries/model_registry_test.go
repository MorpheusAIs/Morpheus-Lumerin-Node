package registries

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestMorRpc_modelAgentId_to_Hex(t *testing.T) {
	modelAgentId := [32]byte{0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		105,
		112,
		102,
		115,
		58,
		47,
		47,
		105,
		112,
		102,
		115,
		97,
		100,
		100,
		114,
		101,
		115,
		115}
	modelAgentIdHex := hex.EncodeToString(modelAgentId[:])
	fmt.Println(modelAgentIdHex)
}
