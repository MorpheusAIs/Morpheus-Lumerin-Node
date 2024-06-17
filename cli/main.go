package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/Lumerin-protocol/Morpheus-Lumerin-Node/api-gateway/client"
	chat "github.com/Lumerin-protocol/Morpheus-Lumerin-Node/cli/chat"

	dotenv "github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

const httpErrorMessage string = "internal error: %v; http status: %v"

var sessionId string

func main() {
	api_host := "http://localhost:8082"
	dotenv.Load(".env")

	if v := os.Getenv("API_HOST"); v != "" {
		api_host = v
	}

	actions := NewActions(client.NewApiGatewayClient(api_host, http.DefaultClient))
	app := &cli.App{
		Usage: "A client to call the Morpheus Lumerin API",
		Commands: []*cli.Command{
			{
				Name:    "getAllowance",
				Aliases: []string{"ga"},
				Usage:   "get allowance",
				Action:  actions.getAllowance,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "spender",
						Required: true,
					},
				},
			},
			{
				Name:    "approveAllowance",
				Aliases: []string{"aa"},
				Usage:   "approve allowance",
				Action:  actions.approveAllowance,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "spender",
						Required: true,
					},
					&cli.Uint64Flag{
						Name:     "amount",
						Required: true,
					},
				},
			},
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
				Usage:   "create chat completions by sending a prompt to the AI engine",
				Action:  actions.createChatCompletions,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "prompt",
						Required: true,
					},
					&cli.StringFlag{
						Name: "messages",
					},
				},
			},
			{
				Name:    "initiateProxySession",
				Aliases: []string{"ips"},
				Usage:   "initiate a proxy session",
				Action:  actions.initiateProxySession,
			},
			{
				Name:    "blockchainProviders",
				Aliases: []string{"bp"},
				Usage:   "list blockchain providers",
				Action:  actions.blockchainProviders,
			},
			{
				Name:    "createBlockchainProvider",
				Aliases: []string{"bpc"},
				Usage:   "create a blockchain provider",
				Action:  actions.createBlockchainProvider,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "address",
						Required: true,
					},
					&cli.Uint64Flag{
						Name:     "stake",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "endpoint",
						Required: true,
					},
				},
			},
			{
				Name:    "blockchainProviderBid",
				Aliases: []string{"bpb"},
				Usage:   "list provider bids",
				Action:  actions.blockchainProvidersBids,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "address",
						Required: true,
					},
					&cli.Int64Flag{
						Name: "offset",
					},
					&cli.UintFlag{
						Name: "limit",
					},
				},
			},
			{
				Name:    "createClockchainProvidersBids",
				Aliases: []string{"cbpb"},
				Usage:   "create provider bid",
				Action:  actions.createBlockchainProviderBid,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "provider",
						Required: true,
					},
					&cli.StringFlag{
						Name: "model",
					},
					&cli.Uint64Flag{
						Name: "pricePerSecond",
					},
				},
			},
			{
				Name:    "blockchainModels",
				Aliases: []string{"bm"},
				Usage:   "list models",
				Action:  actions.blockchainModels,
			},
			{
				Name:    "openBlockchainSession",
				Aliases: []string{"obs"},
				Usage:   "open a blockchain session",
				Action:  actions.openBlockchainSession,
			},
			{
				Name:    "closeBlockchainSession",
				Aliases: []string{"cbs"},
				Usage:   "close a blockchain session",
				Action:  actions.closeBlockchainSession,
			},
			{
				Name:    "createAndStreamChatCompletions",
				Aliases: []string{"csc"},
				Usage:   "create and stream chat completions",
				Action:  actions.createAndStreamChatCompletions,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "prompt",
						Required: true,
					},
					&cli.StringFlag{
						Name: "messages",
					},
				},
			},
			{
				Name:    "createAndStreamSessionChatCompletions",
				Aliases: []string{"cssc"},
				Usage:   "create and stream session chat completions",
				Action:  actions.createChatCompletions,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "prompt",
						Required: true,
					},
					&cli.StringFlag{
						Name: "messages",
					},
				},
			},
			{
				Name:    "chat",
				Aliases: []string{},
				Usage:   "open interactive chat client",
				Action:  actions.startChat,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "sessionId",
						Required: false,
					},
				},
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

func (a *actions) startChat(cCtx *cli.Context) error {

	chat.Run()

	return nil
}

func (a *actions) getAllowance(cCtx *cli.Context) error {
	spender := cCtx.String("spender")
	res, err := a.client.GetAllowance(cCtx.Context, spender)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func (a *actions) approveAllowance(cCtx *cli.Context) error {
	spender := cCtx.String("spender")
	amount := cCtx.Uint64("amount")
	res, err := a.client.ApproveAllowance(cCtx.Context, spender, amount)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
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
	var messages []client.ChatCompletionMessage
	json.Unmarshal([]byte(cCtx.String("messages")), &messages)

	completion, err := a.client.Prompt(cCtx.Context, prompt, messages)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(completion)
	fmt.Println(string(jsonData))
	return nil
}

// todo: retrieve session id
func (a *actions) createAndStreamChatCompletions(cCtx *cli.Context) error {
	prompt := cCtx.String("prompt")
	var messages []*client.ChatCompletionMessage
	json.Unmarshal([]byte(cCtx.String("messages")), &messages)

	completion, err := a.client.PromptStream(cCtx.Context, prompt, "", func(msg *client.ChatCompletionStreamResponse) error {
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

	for _, item := range providers["providers"].([]interface{}) {
		provider := item.(map[string]interface{})
		// fmt.Println(provider)
		// fmt.Println(reflect.TypeOf(provider))
		fmt.Println(provider["Address"], " - ", provider["Endpoint"])
	}

	// jsonData, err := json.Marshal(providers)

	// if err != nil {
	// 	return err
	// }

	return nil
}

func (a *actions) createBlockchainProvider(cCtx *cli.Context) error {
	address := cCtx.String("address")
	stake := cCtx.Uint64("stake")
	endpoint := cCtx.String("endpoint")

	providers, err := a.client.CreateNewProvider(cCtx.Context, address, stake, endpoint)

	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(providers)
	fmt.Println(string(jsonData))
	return nil
}

func (a *actions) createBlockchainProviderBid(cCtx *cli.Context) error {
	provider := cCtx.String("provider")
	model := cCtx.String("model")
	pricePerSecond := cCtx.Uint64("pricePerSecond")

	_, err := a.client.CreateNewProviderBid(cCtx.Context, provider, model, pricePerSecond)

	if err != nil {
		return err
	}

	fmt.Println("bid created for provider ", provider)
	return nil
}

func (a *actions) blockchainProvidersBids(cCtx *cli.Context) error {
	address := cCtx.String("address")
	offset := cCtx.Int64("offset")
	limit := cCtx.Uint("limit")

	bids, err := a.client.GetBidsByProvider(cCtx.Context, address, big.NewInt(offset), uint8(limit))
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

	req := &client.SessionRequest{}

	session, err := a.client.OpenSession(req, cCtx.Context)

	if err != nil {
		return err
	}
	// TODO: Output a message indicating the blockchain session was opened and showing relevant data for the session
	// jsonData, err := json.Marshal(bids)
	fmt.Println(session)
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
