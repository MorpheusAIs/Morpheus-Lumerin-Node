package common

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"

// DefaultSessionDuration is the session length used by `chat` when --duration
// is not given, in seconds.
//
// Opening a session escrows stake for its whole duration up front, so this
// default decides whether opening one succeeds at all. At a typical bid price
// of ~2.8e13 wei/second an hour requires escrowing roughly 100 MOR, which is
// more than most consumer wallets hold, so the previous hardcoded 3600 failed
// with an insufficient-balance error for many users. The stake is not spent --
// unused stake returns on close -- but it has to be available up front. Ten
// minutes is enough to try a model out, and --duration raises it when wanted.
const DefaultSessionDuration int64 = 600

type Options struct {
	Edit         bool
	List         bool
	Remove       bool
	Version      bool
	Session      string
	Model        string
	PrivateKey   string
	LocalModels  []interface{}
	RemoteModels []interface{}

	// SessionDuration is the length in seconds of a session opened from the
	// model picker. Zero means DefaultSessionDuration.
	SessionDuration int64

	UseLocalModel bool
	Client        *client.ApiGatewayClient
}

// Duration returns the configured session duration, or the default when unset.
func (o *Options) Duration() int64 {
	if o.SessionDuration <= 0 {
		return DefaultSessionDuration
	}
	return o.SessionDuration
}
