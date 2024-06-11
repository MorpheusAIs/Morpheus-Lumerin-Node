package main

import (
	"flag"
	"os"

	"github.com/dwisiswant0/chatgptui/chat"
	"github.com/dwisiswant0/chatgptui/common"
	"github.com/dwisiswant0/chatgptui/config"
	"github.com/dwisiswant0/chatgptui/util"
)

func init() {
	flag.BoolVar(&opt.Edit, "e", false, "Edit configuration file")
	flag.BoolVar(&opt.Edit, "edit", false, "Edit configuration file")

	flag.BoolVar(&opt.List, "l", false, "List all supported OpenAI model")
	flag.BoolVar(&opt.List, "list", false, "List all supported OpenAI model")

	flag.BoolVar(&opt.Remove, "rm", false, "Remove configuration file")

	flag.BoolVar(&opt.Version, "V", false, "Show current version")
	flag.BoolVar(&opt.Version, "version", false, "Show current version")

	openAiBaseUrl := os.Getenv("OPENAI_BASE_URL")

	if openAiBaseUrl == "" {
		os.Setenv("OPENAI_BASE_URL","http://localhost:8082/v1")
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

func main() {
	cfgPath := common.GetConfigPath()

	cfg, err := config.Load(cfgPath)
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
