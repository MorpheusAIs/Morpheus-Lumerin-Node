package common

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/client"

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

	UseLocalModel bool
	Client        *client.ApiGatewayClient
}
