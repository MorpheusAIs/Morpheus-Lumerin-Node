package lib

import (
	"fmt"
)

type customError struct {
	err      error
	childErr error
}

func (e *customError) Error() string {
	var (
		parentStr = "<nil>"
		childStr  = "<nil>"
	)
	if e.err != nil {
		parentStr = e.err.Error()
	}
	if e.childErr != nil {
		childStr = e.childErr.Error()
	}
	return fmt.Sprintf("%s: %s", parentStr, childStr)
}

func (e *customError) Unwrap() error {
	return e.childErr
}

func (e *customError) Is(target error) bool {
	return target == e.err
}

// WrapError allows to use reuse predefined error objects and add dynamic data
// allowing to create error chains
//
//	Example:
//	var ErrEnvParse = errors.New("cannot parse env variable")
//	var err = fmt.Errorf("IS_BUYER should be boolean")
//	wrappedErr := WrapError(ErrEnvParse, err)
//
//	errors.Is(wrappedErr, ErrEnvParse) 	// true
//	errors.Unwrap(wrappedErr).Is(err) 	// true
//	wrappedErr.Error()	// "cannot parse env variable: IS_BUYER should be boolean"
func WrapError(parent error, child error) error {
	return &customError{
		err:      parent,
		childErr: child,
	}
}
