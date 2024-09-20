package lib

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type EVMError struct {
	Abi   abi.Error
	Cause error
	Args  interface{}
}

func (e EVMError) Error() string {
	idBytes := e.Abi.ID.Bytes()
	if len(idBytes) > 4 {
		idBytes = idBytes[:4]
	}
	return "EVM error: " + e.Abi.Sig + " " + common.BytesToHash(idBytes).Hex()
}

// TryConvertGethError attempts to convert geth error to an EVMError, otherwise just returns original error
func TryConvertGethError(err error, contractMeta *bind.MetaData) error {
	evmErr, ok := ConvertGethError(err, contractMeta)
	if !ok {
		return err
	}
	return evmErr
}

// ConvertGethError converts a geth error to an EVMError with exposed error signature and arguments
func ConvertGethError(err error, contractMeta *bind.MetaData) (*EVMError, bool) {
	errData, ok := ExtractGETHErrorData(err)
	fmt.Println(errData, ok)
	if !ok {
		return nil, false
	}

	abiError, args, ok := CastErrorData(errData, contractMeta)
	fmt.Println(abiError, args, ok)
	if !ok {
		return nil, false
	}

	return &EVMError{
		Abi:   abiError,
		Args:  args,
		Cause: err,
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
func CastErrorData(errData []byte, contractMetadata *bind.MetaData) (abi.Error, interface{}, bool) {
	if abiData, err := contractMetadata.GetAbi(); err == nil {
		for _, abiError := range abiData.Errors {
			args, err := abiError.Unpack(errData)
			if err == nil {
				return abiError, args, true
			}
		}
	}
	return abi.Error{}, nil, false
}
