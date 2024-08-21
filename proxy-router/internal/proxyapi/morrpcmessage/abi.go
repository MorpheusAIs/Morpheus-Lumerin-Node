package morrpcmesssage

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"

var approvalAbi = []lib.AbiParameter{
	{Type: "bytes32"}, // bidID
	{Type: "uint256"}, // chainID
	{Type: "address"}, // user
	{Type: "uint128"}, // timestamp
}

var sessionReportAbi = []lib.AbiParameter{
	{Type: "bytes32"}, // sessionID
	{Type: "uint256"}, // chainID
	{Type: "uint128"}, // timestamp
	{Type: "uint32"},  // tpsScaled1000
	{Type: "uint32"},  // ttftMs
}
