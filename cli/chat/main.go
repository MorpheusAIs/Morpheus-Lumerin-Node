package mainchat

import (
	chat "github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/chat"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/common"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/cli/chat/util"
)

func Run(opt *common.Options) {
	m = chat.New(opt)
	util.RunProgram(m)
}
