package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/api-gateway/client"

	"github.com/urfave/cli/v2"
)

const httpErrorMessage string = "internal error: %v; http status: %v"

func main() {
	actions := NewActions(client.NewApiGatewayClient("http://localhost:8080", http.DefaultClient))
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
				Subcommands: []*cli.Command{
					{
						Name:    "create",
						Aliases: []string{"c"},
						Usage:   "blockchainProviders create",
						Action:  actions.createBlockchainProvider,
					},
				}
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
			{
				Name:    "createAndStreamChatCompletions",
				Aliases: []string{"csc", "streamlocal"},
				Usage:   "create a chat completion by sending a prompt to the ai engine",
				Action:  actions.createAndStreamChatCompletions,
			},
			{
				Name:    "createAndStreamSessionChatCompletions",
				Aliases: []string{"cssc"},
				Usage:   "create a chat completion by sending a prompt to the ai engine",
				Action:  actions.createChatCompletions,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type actions struct {
	client *client.ApiGatewayClient
}

func NewActions(c *client.ApiGatewayClient) *actions {
	return &actions{client: c}
}

func (a *actions) healthcheck(cCtx *cli.Context) error {
	res, err := a.client.HealthCheck(cCtx.Context)
	if err != nil {
		return err
	}
	// Output the result of the healthcheck
	jsonData, err := json.Marshal(res)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) proxyRouterConfig(cCtx *cli.Context) error {
	config, err := a.client.GetProxyRouterConfig(cCtx.Context)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(config)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) proxyRouterFiles(cCtx *cli.Context) error {
	files, err := a.client.GetProxyRouterFiles(cCtx.Context)

	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(files)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) createChatCompletions(cCtx *cli.Context) error {
	prompt := cCtx.String("prompt")

	//TODO: handle chat history
	_ = cCtx.StringSlice("messages")

	completion, err := a.client.Prompt(cCtx.Context, prompt, []client.ChatCompletionMessage{})
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(completion)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) createAndStreamChatCompletions(cCtx *cli.Context) error {
	prompt := cCtx.String("prompt")
	messages := cCtx.Generic("messages").([]*client.ChatCompletionMessage)

	completion, err := a.client.PromptStream(cCtx.Context, prompt, messages, func(msg client.ChatCompletionStreamResponse) error {
		fmt.Println(msg)
		return nil
	})
	
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(completion)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) createSessionChatCompletions(cCtx *cli.Context) error {
	prompt := cCtx.String("prompt")

	messages := cCtx.StringSlice("messages")

	completion, err := a.client.SessionPrompt(cCtx.Context, prompt, messages)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(completion)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) initiateProxySession(cCtx *cli.Context) error {
	session, err := a.client.InitiateSession(cCtx.Context)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(session)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) blockchainProviders(cCtx *cli.Context) error {
	providers, err := a.client.GetAllProviders(cCtx.Context)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(providers)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) createBlockchainProvider(cCtx *cli.Context) error {
	//TODO: handle provider fields
	providers, err := a.client.createProvider(cCtx.Context)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(providers)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) blockchainProvidersBids(cCtx *cli.Context) error {

	providerAddress := cCtx.String("providerAddress")

	offset := cCtx.Int64("offset")

	limit := cCtx.Uint("limit")

	bids, err := a.client.GetBidsByProvider(cCtx.Context, providerAddress, big.NewInt(offset), uint8(limit))
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(bids)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) blockchainModels(cCtx *cli.Context) error {
	models, err := a.client.GetAllModels(cCtx.Context)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(models)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) openBlockchainSession(cCtx *cli.Context) error {
	err := a.client.OpenSession(cCtx.Context)
	if err != nil {
		return err
	}
	// TODO: Output a message indicating the blockchain session was opened and showing relevant data for the session
	// jsonData, err := json.Marshal(bids)
	// fmt.Println(string(jsonData))
	return nil
}

func (a *actions) closeBlockchainSession(cCtx *cli.Context) error {
	err := a.client.CloseSession(cCtx.Context)
	if err != nil {
		return err
	}

	// TODO: Output a message indicating the blockchain session was closed and showing relevant data for the session
	// jsonData, err := json.Marshal(bids)
	// fmt.Println(string(jsonData))
	return nil
}
