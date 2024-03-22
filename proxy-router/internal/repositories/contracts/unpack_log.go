package contracts

import (
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	ErrNoEventSignature       = errors.New("no event signature")
	ErrEventSignatureMismatch = errors.New("event signature mismatch")
)

// UnpackLog unpacks the log into the provided struct, copied from go-ethereum implementation
func UnpackLog(out interface{}, event string, log types.Log, ab *abi.ABI) error {
	// Anonymous events are not supported.
	if len(log.Topics) == 0 {
		return ErrNoEventSignature
	}
	if log.Topics[0] != ab.Events[event].ID {
		return ErrEventSignatureMismatch
	}
	if len(log.Data) > 0 {
		if err := ab.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range ab.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}
