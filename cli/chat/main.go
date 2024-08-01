package mainchat

import (
	"flag"
	"fmt"
	"os"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/chat"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/util"
)

func init() {
	flag.BoolVar(&opt.Edit, "e", false, "Edit configuration file")
	flag.BoolVar(&opt.Edit, "edit", false, "Edit configuration file")

	flag.BoolVar(&opt.List, "l", false, "List all supported OpenAI model")
	flag.BoolVar(&opt.List, "list", false, "List all supported OpenAI model")

	flag.BoolVar(&opt.Remove, "rm", false, "Remove configuration file")

	flag.BoolVar(&opt.Version, "v", false, "Show current version")
	flag.BoolVar(&opt.Version, "version", false, "Show current version")

	openAiBaseUrl := os.Getenv("OPENAI_BASE_URL")

	if openAiBaseUrl == "" {
		os.Setenv("OPENAI_BASE_URL", "http://localhost:8082/v1")
	}

	flag.Usage = func() {
		showBanner()
		showUsage()
	}
	flag.Parse()

	switch {
	case opt.List:
		listAllModels()
	case opt.Remove:
		removeConfig()
	case opt.Version:
		showVersion()
	}

	// if opt.List {
	// 	listAllModels()
	// }

	// if opt.Remove {
	// 	removeConfig()
	// }

	// if opt.Version {
	// 	showVersion()
	// }
}

func Run(opt *common.Options) {

	switch {
	case opt.List:
		listAllModels()
	case opt.Remove:
		removeConfig()
	case opt.Version:
		showVersion()
	default:

		cfgPath := common.GetConfigPath()
fmt.Println("cfgPath: ", cfgPath)
		cfg, err := config.Load(cfgPath)

		fmt.Printf("cfg: %+v\n", cfg)
		fmt.Printf("err: %v\n", err)
		if opt.Session != "" {
			cfg.SessionId = opt.Session
		}

		if opt.Model != "" {
			cfg.ModelId = opt.Model
		}

		if opt.PrivateKey != "" {
			cfg.WalletKey = opt.PrivateKey
		}
		fmt.Printf("cfg with options: %+v\n", cfg)

		if err == nil {
			m = chat.New(cfg)

			if opt.Edit {
				m = config.New(cfg)
			}
		} else {
			m = config.New()
		}

		util.RunProgram(m)
	}
}
