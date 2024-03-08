package proxy

import "errors"

var (
	ErrDest   = errors.New("destination connection error")
	ErrSource = errors.New("source connection error")
)
