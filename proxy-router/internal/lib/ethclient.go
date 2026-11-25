package lib

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/delegatestaking"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/delegatestakingfunding"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/delegation"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/lumerintoken"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/marketplace"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/modelregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/morpheustoken"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/multicall3"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerpayouts"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Meta interface {
	GetAbi() (*abi.ABI, error)
}

var allContractsMeta = []Meta{
	marketplace.MarketplaceMetaData,
	modelregistry.ModelRegistryMetaData,
	providerregistry.ProviderRegistryMetaData,
	sessionrouter.SessionRouterMetaData,
	morpheustoken.MorpheusTokenMetaData,
	lumerintoken.LumerinTokenMetaData,
	multicall3.Multicall3MetaData,
	delegation.DelegationMetaData,
	providerpayouts.ProviderPayoutsMetaData,
	delegatestaking.DelegateStakingMetaData,
	delegatestakingfunding.DelegateStakingFundingMetaData,
}

type EVMError struct {
	Abi   abi.Error
	Cause error
	Args  interface{}
}

func (e EVMError) Error() string {
	return fmt.Sprintf("EVM error: %s %+v", e.Abi.Sig, e.Args)
}

// Implement As() method to check if EVMError can be converted to another type.
func (e EVMError) As(target interface{}) bool {
	// Ensure that the target is a pointer.
	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		// As target should be a pointer
		return false
	}

	switch v := target.(type) {
	case *EVMError:
		*v = e // Assign the concrete EVMError to the target
		return true
	default:
		return false
	}
}

// TryConvertGethError attempts to convert geth error to an EVMError, otherwise just returns original error
func TryConvertGethError(err error) error {
	evmErr, ok := ConvertGethError(err, allContractsMeta)
	if !ok {
		return err
	}
	return evmErr
}

// diamondSelectorNotRegistered is the reason the diamond reverts with when no
// facet serves the called function.
const diamondSelectorNotRegistered = "Diamond: selector is not registered"

// IsSelectorNotRegistered reports whether err is the diamond's revert for a
// function no facet serves: a getter this diamond does not route, or a Stage 1
// function on a diamond without the Stage 1 cut. The revert data is decoded
// when the node returns it, because not every node repeats the reason in the
// error message.
func IsSelectorNotRegistered(err error) bool {
	if err == nil {
		return false
	}
	var dataErr interface {
		error
		ErrorData() interface{}
	}
	if errors.As(err, &dataErr) {
		if data, ok := ExtractGETHErrorData(dataErr); ok {
			if reason, unpackErr := abi.UnpackRevert(data); unpackErr == nil {
				return reason == diamondSelectorNotRegistered
			}
		}
	}
	return strings.Contains(err.Error(), "selector is not registered")
}

// ConvertGethError converts a geth error to an EVMError with exposed error signature and arguments
func ConvertGethError(gethErr error, contractMeta []Meta) (*EVMError, bool) {
	errData, ok := ExtractGETHErrorData(gethErr)
	if !ok {
		return nil, false
	}

	abis, err := getAbi(contractMeta)
	if err != nil {
		return nil, false
	}

	abiError, args, ok := CastErrorData(errData, abis)
	if !ok {
		return nil, false
	}

	return &EVMError{
		Abi:   abiError,
		Args:  args,
		Cause: gethErr,
	}, true
}

// ExtractGETHErrorData extracts the error data from an unexproted geth error
func ExtractGETHErrorData(err error) ([]byte, bool) {
	asErr, ok := err.(interface{ ErrorData() interface{} })
	if !ok {
		return nil, false
	}
	errDataHex, ok := asErr.ErrorData().(string)
	if !ok {
		return nil, false
	}
	errDataBytes := common.FromHex(errDataHex)
	if len(errDataBytes) < 4 {
		return nil, false
	}
	return errDataBytes, true
}

// CastErrorData casts the error data to the appropriate error type
func CastErrorData(errData []byte, abis []*abi.ABI) (abi.Error, interface{}, bool) {
	for _, abiData := range abis {
		for _, abiError := range abiData.Errors {
			args, err := abiError.Unpack(errData)
			if err == nil {
				return abiError, args, true
			}
		}
	}
	return abi.Error{}, nil, false
}

// DecodeInput decodes the input data of a transaction
func DecodeInput(rawInput []byte) (methodName string, entries []InputEntry, err error) {
	abi, err := getAbi(allContractsMeta)
	if err != nil {
		return "", nil, err
	}

	for _, a := range abi {
		// find method across all ABIs
		method, err := a.MethodById(rawInput)
		if err != nil {
			continue
		}

		// decode input
		args := make(map[string]interface{})
		err = method.Inputs.UnpackIntoMap(args, rawInput[4:])
		if err != nil {
			return "", nil, err
		}

		// remap to our format
		entries := make([]InputEntry, len(args))
		for k, v := range method.Inputs {
			entries[k] = InputEntry{
				Key:   v.Name,
				Type:  v.Type.String(),
				Value: args[v.Name],
			}
		}

		return method.RawName, entries, nil
	}

	return "", nil, fmt.Errorf("method not found")
}

type InputEntry struct {
	Key   string
	Type  string
	Value interface{}
}

func getAbi(meta []Meta) ([]*abi.ABI, error) {
	abis := make([]*abi.ABI, len(meta))
	for i, m := range meta {
		abi, err := m.GetAbi()
		if err != nil {
			return nil, err
		}
		abis[i] = abi
	}
	return abis, nil
}
