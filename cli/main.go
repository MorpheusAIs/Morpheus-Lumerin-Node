package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/api-gateway/client"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "healthcheck",
				Aliases: []string{"he"},
				Usage:   "check application health",
				Action:  actions.healthcheck,
			},
			{
				Name:    "proxyRouterConfig",
				Aliases: []string{"prc"},
				Usage:   "view proxy router config",
				Action:  actions.proxyRouterConfig,
			},
			{
				Name:    "proxyRouterFiles",
				Aliases: []string{"prf"},
				Usage:   "get the files associated with the proxy router pid",
				Action:  actions.proxyRouterFiles,
			},
			{
				Name:    "createChatCompletions",
				Aliases: []string{"ccc"},
				Usage:   "create a chat completion by sending a prompt to the ai engine",
				Action:  actions.createChatCompletions,
			},
			{
				Name:    "initiateProxySession",
				Aliases: []string{"ips"},
				Usage:   "",
				Action:  actions.initiateProxySession,
			},
			{
				Name:    "blockchainProviders",
				Aliases: []string{"bp"},
				Usage:   "",
				Action:  actions.blockchainProviders,
			},
			{
				Name:    "blockchainProviders",
				Aliases: []string{"bp"},
				Usage:   "",
				Action:  actions.blockchainProviders,
			},
			{
				Name:    "blockchainProvidersBids",
				Aliases: []string{"bpb"},
				Usage:   "",
				Action:  actions.blockchainProvidersBids,
			},
			{
				Name:    "blockchainModels",
				Aliases: []string{"bm"},
				Usage:   "",
				Action:  actions.blockchainModels,
			},
			{
				Name:    "openBlockchainSession",
				Aliases: []string{"open blockchain session"},
				Usage:   "",
				Action:  actions.openBlockchainSession,
			},
			{
				Name:    "closeBlockchainSession",
				Aliases: []string{"cbs"},
				Usage:   "",
				Action:  actions.closeBlockchainSession,
			},
		},
}

var actions = &actions{}

	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type actions struct {
	client *client.Client
}

func NewActions(c *client.Client) *actions {
	return &actions{client: c}
}
func (a *actions) healthcheck(cCtx *cli.Context) error {
	res, err := a.client.Healthcheck()
	if err != nil {
		return err
	}
	fmt.Println("healthcheck:", res) // Output the result of the healthcheck
	return nil
}

func (a *actions) proxyRouterConfig(cCtx *cli.Context) error {
	config, err := a.client.GetProxyRouterConfig()
	if err != nil {
		return err
	}
	fmt.Println("proxy router config:", config) // Output the proxy router configuration
	return nil
}

func (a *actions) proxyRouterFiles(cCtx *cli.Context) error {
	files, err := a.client.GetProxyRouterFiles()
	if err != nil {
		return err
	}
	fmt.Println("proxy router files:", files) // Output the proxy router files
	return nil
}

func (a *actions) createChatCompletions(cCtx *cli.Context) error {
	completion, err := a.client.CreateChatCompletion()
	if err != nil {
		return err
	}
	fmt.Println("chat completion:", completion) // Output the chat completion
	return nil
}

func (a *actions) initiateProxySession(cCtx *cli.Context) error {
	session, err := a.client.InitiateProxySession()
	if err != nil {
		return err
	}
	fmt.Println("proxy session:", session) // Output the proxy session details
	return nil
}

func (a *actions) blockchainProviders(cCtx *cli.Context) error {
	providers, err := a.client.GetBlockchainProviders()
	if err != nil {
		return err
	}
	fmt.Println("blockchain providers:", providers) // Output the blockchain providers
	return nil
}

func (a *actions) blockchainProvidersBids(cCtx *cli.Context) error {
	bids, err := a.client.GetBlockchainProvidersBids()
	if err != nil {
		return err
	}
	fmt.Println("blockchain providers bids:", bids) // Output the blockchain providers' bids
	return nil
}

func (a *actions) blockchainModels(cCtx *cli.Context) error {
	models, err := a.client.GetBlockchainModels()
	if err != nil {
		return err
	}
	fmt.Println("blockchain models:", models) // Output the blockchain models
	return nil
}

func (a *actions) openBlockchainSession(cCtx *cli.Context) error {
	session, err := a.client.OpenBlockchainSession()
	if err != nil {
		return err
	}
	fmt.Println("blockchain session:", session) // Output the blockchain session details
	return nil
}

func (a *actions) closeBlockchainSession(cCtx *cli.Context) error {
	err := a.client.CloseBlockchainSession()
	if err != nil {
		return err
	}
	fmt.Println("blockchain session closed") // Output a message indicating the blockchain session was closed
	return nil
}
